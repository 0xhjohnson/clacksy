package models

import "errors"

var (
	ErrDuplicateEmail     = errors.New("models: duplicate email")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateUsername  = errors.New("models: duplicate username")
)
