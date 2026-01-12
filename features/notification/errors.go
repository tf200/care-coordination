package notification

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInternal       = errors.New("internal server error")
	ErrNotFound       = errors.New("notification not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidToken   = errors.New("invalid or expired token")
	ErrMissingToken   = errors.New("missing authentication token")
	ErrInvalidTicket  = errors.New("invalid or expired ticket")
)
