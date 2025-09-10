package models

import "errors"

var (
	ErrSignatureNotFound      = errors.New("signature not found")
	ErrSignatureAlreadyExists = errors.New("signature already exists")
	ErrInvalidUser            = errors.New("invalid user")
	ErrInvalidDocument        = errors.New("invalid document ID")
	ErrDatabaseConnection     = errors.New("database connection error")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrDomainNotAllowed       = errors.New("domain not allowed")
)
