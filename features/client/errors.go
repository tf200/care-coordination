package client

import "errors"

var (
	ErrInvalidRequest           = errors.New("invalid request")
	ErrIntakeFormNotFound       = errors.New("intake form not found")
	ErrRegistrationFormNotFound = errors.New("registration form not found")
	ErrFailedToCreateClient     = errors.New("failed to create client")
	ErrInternal                 = errors.New("internal server error")
)
