package domain

import "errors"

var (
	ErrNotFound      = errors.New("metadata not found")
	ErrInvalidURL    = errors.New("invalid url")
	ErrInternalError = errors.New("internal server error")
)
