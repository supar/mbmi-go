package main

import (
	crand "crypto/rand"
	"math/rand"
	"time"
)

const (
	_lower        = "abcdefghijklmnopqrstuvwxyz"
	_upper        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	_digits       = "0123456789"
	_symbols      = "~!@#$%^&*()_+-={}|[]:<>?,./"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringId(n int) string {
	var (
		letterBytes = _lower + _upper + _digits
		b           = make([]byte, n)
	)

	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func createSecret(n int, no_upper, no_digits, no_symbols bool) (str string, err error) {
	var (
		letters = _lower
		output  = make([]byte, n)
		random  = make([]byte, n)
	)

	if _, err = crand.Read(random); err != nil {
		return "", err
	}

	if !no_upper {
		letters += _upper
	}

	if !no_digits {
		letters += _digits
	}

	if !no_symbols {
		letters += _symbols
	}

	for all, i := len(letters), n-1; i >= 0; {
		output[i] = letters[uint8(random[i])%uint8(all)]
		i--
	}

	return string(output), err
}
