package locTransfer

import (
	"errors"
)

var (
	// ErrInvalidRequest is returned when the request is invalid.
	ErrInvalidRequest = errors.New("invalid request")

	// ErrClientNotFound is returned when the client is not found in the database.
	ErrClientNotFound = errors.New("client not found")

	// ErrInternal is returned when an internal error occurs.
	ErrInternal = errors.New("internal server error")
)
