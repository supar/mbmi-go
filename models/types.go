package models

import (
	"errors"
	"strings"
)

type Email string
type Boolean bool

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

func (bit *Boolean) UnmarshalJSON(data []byte) error {
	var value = string(data)

	if value == "1" || value == "true" {
		*bit = true
	} else if value == "0" || value == "false" {
		*bit = false
	} else {
		return errors.New("Boolean unmarshal error: invalid input `" + value + "`")
	}

	return nil
}

func (scanner *Boolean) Scan(src interface{}) error {
	switch src.(type) {
	case int64:
		if v, ok := src.(int64); ok && v > 0 {
			*scanner = Boolean(true)
		}

	case bool:
		if v, ok := src.(bool); ok {
			*scanner = Boolean(v)
		}
	}

	return nil
}

func (driver Boolean) Value() (value interface{}, err error) {
	if driver {
		return int64(1), nil
	}
	return int64(0), nil
}
