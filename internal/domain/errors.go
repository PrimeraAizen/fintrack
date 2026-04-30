package domain

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrConflict            = errors.New("resource conflict")
	ErrInvalidInput        = errors.New("invalid input")
	ErrValidation          = errors.New("validation failed")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid or expired token")
)
