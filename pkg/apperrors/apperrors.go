package apperrors

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input provided")

	ErrUnauthorized = errors.New("unauthorized: invalid or missing session")

	ErrForbidden = errors.New("forbidden: you do not have permission to perform this action")

	ErrNotFound = errors.New("resource not found")

	ErrAlreadyExists = errors.New("resource already exists")
	
	ErrInternal = errors.New("internal server error")
)