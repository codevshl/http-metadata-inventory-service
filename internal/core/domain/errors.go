package domain

import "github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"

var (
	ErrNotFound      = apperror.New(apperror.ErrNotFound, "metadata not found")
	ErrInvalidURL    = apperror.New(apperror.ErrValidation, "invalid url")
	ErrInternalError = apperror.New(apperror.ErrInternal, "internal server error")
)
