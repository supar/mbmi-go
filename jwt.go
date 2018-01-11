package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"mbmi-go/models"
)

// JSON web token object
type Token struct {
	JWT      string `json:"jwt"`
	secret   []byte
	identity *models.User
	t        *jwt.Token
}

type IdentityIface interface {
	Identity() int64
}

type TokenClaims struct {
	Uid int64 `json:"uid"`
	jwt.StandardClaims
}

func NewClaims() TokenClaims {
	return TokenClaims{
		StandardClaims: jwt.StandardClaims{},
	}
}

func NewToken(secret []byte) (t *Token) {
	return &Token{
		secret: secret,
	}
}

func (s *Token) Sign(claims TokenClaims) *Token {
	s.t = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s.JWT, _ = s.t.SignedString(s.secret)

	return s
}

func (s *Token) Parse() (err error) {
	s.t, err = jwt.ParseWithClaims(s.JWT, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return s.secret, nil
	})

	return
}

func (s *Token) Identity() int64 {
	if s.t != nil {
		return s.t.Claims.(*TokenClaims).Uid
	}

	return -1
}

func (s *Token) Valid() bool {
	if s.t != nil {
		return s.t.Valid
	}

	return false
}
