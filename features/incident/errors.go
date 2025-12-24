package incident

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInternal       = errors.New("internal server error")
	ErrNotFound       = errors.New("incident not found")
)
