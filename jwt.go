package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"mbmi-go/models"
)

// Token represent user authorization data
type Token struct {
	JWT      string `json:"jwt"`
	secret   []byte
	identity *models.User
	t        *jwt.Token
}

// IdentityIface is used to work with Token data
type IdentityIface interface {
	Identity() int64
	Issuer() string
	Valid() bool
	Subject() string
}

// TokenClaims represents extention for the standard claims
// from JWT package
type TokenClaims struct {
	uid int64 `json:"uid"`
	jwt.StandardClaims
}

// NewClaims returns new token claims
func NewClaims(uid int64, subject string) TokenClaims {
	return TokenClaims{
		uid: uid,
		StandardClaims: jwt.StandardClaims{
			Subject: subject,
		},
	}
}

// NewToken returns new token
func NewToken(secret []byte) (t *Token) {
	return &Token{
		secret: secret,
	}
}

// Sign token
func (s *Token) Sign(claims TokenClaims) *Token {
	s.t = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s.JWT, _ = s.t.SignedString(s.secret)

	return s
}

// Parse JWT string with TokenClaims and validate token
func (s *Token) Parse() (err error) {
	s.t, err = jwt.ParseWithClaims(s.JWT, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return s.secret, nil
	})

	return
}

// Identity returns user id
func (s *Token) Identity() int64 {
	if s.t != nil {
		return s.t.Claims.(*TokenClaims).uid
	}

	return -1
}

// Valid returns token state: true if token valid
func (s *Token) Valid() bool {
	if s.t != nil {
		return s.t.Valid
	}

	return false
}

// Issuer returns issue value
func (s *Token) Issuer() string {
	return s.t.Claims.(*TokenClaims).Issuer
}

// Subject returns token subject
func (s *Token) Subject() string {
	if s.t != nil {
		return s.t.Claims.(*TokenClaims).Subject
	}
	return ""
}
