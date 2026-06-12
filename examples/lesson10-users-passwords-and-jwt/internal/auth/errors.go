package auth

import "errors"

var (
	ErrValidation   = errors.New("validation error")
	ErrConflict     = errors.New("resource conflict")
	ErrUnauthorized = errors.New("unauthorized")
)
