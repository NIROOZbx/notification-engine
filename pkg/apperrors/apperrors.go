package apperrors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// sentinel errors — for simple cases
var (
	ErrInvalidInput  = errors.New("invalid input provided")
	ErrUnauthorized  = errors.New("unauthorized: invalid or missing session")
	ErrForbidden     = errors.New("forbidden: you do not have permission to perform this action")
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInternal      = errors.New("internal server error")
)

type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

func NewNotFoundError(resource string) error {
	return &NotFoundError{Resource: resource}
}

type AlreadyExistsError struct {
	Resource string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", e.Resource)
}

func NewAlreadyExistsError(resource string) error {
	return &AlreadyExistsError{Resource: resource}
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
