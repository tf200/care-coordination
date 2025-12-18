package employee

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInternal       = errors.New("internal server error")
	ErrUnauthorized   = errors.New("unauthorized")
)
