package auth

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid_request")
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrInvalidToken       = errors.New("invalid_token")
	ErrInvalidMFACode     = errors.New("invalid_mfa_code")
	ErrMFANotSetup        = errors.New("mfa_not_setup")
	ErrMFAAlreadyEnabled  = errors.New("mfa_already_enabled")
	ErrInternal           = errors.New("internal")
)
