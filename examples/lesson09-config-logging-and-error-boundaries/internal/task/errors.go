package task

import (
	"errors"
	"fmt"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("task not found")
)

type DomainError struct {
	Kind    error
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Kind
}

func validationError(format string, args ...any) error {
	return &DomainError{Kind: ErrValidation, Message: fmt.Sprintf(format, args...)}
}

func notFoundError(format string, args ...any) error {
	return &DomainError{Kind: ErrNotFound, Message: fmt.Sprintf(format, args...)}
}

func publicMessage(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Message
	}

	return "unexpected error"
}
