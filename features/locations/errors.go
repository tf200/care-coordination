package locations

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInternal       = errors.New("internal server error")
	ErrNotFound       = errors.New("location not found")
)
