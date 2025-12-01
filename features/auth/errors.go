package auth

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid_request")
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrInvalidToken       = errors.New("invalid_token")
	ErrInternal           = errors.New("internal")
)
