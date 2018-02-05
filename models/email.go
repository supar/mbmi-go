package models

import (
	"errors"
	"strings"
)

type Email string

func (e Email) Split() (login, domain string, err error) {
	var (
		email = strings.Trim(string(e), " ")
		parts = strings.Split(email, "@")
	)

	if l := len(parts); l != 2 {
		err = errors.New("Invalid email format")
		return
	}

	// Simple domain validation
	if strings.ContainsAny(parts[1], " ,!_@#%^&*()[]}{/|\\?`") {
		err = errors.New("Invalid email format")
		return
	}

	return parts[0], parts[1], nil
}
