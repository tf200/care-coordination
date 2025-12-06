package client

import "errors"

var (
	ErrInvalidRequest            = errors.New("invalid request")
	ErrIntakeFormNotFound        = errors.New("intake form not found")
	ErrRegistrationFormNotFound  = errors.New("registration form not found")
	ErrFailedToCreateClient      = errors.New("failed to create client")
	ErrInternal                  = errors.New("internal server error")
	ErrClientNotFound            = errors.New("client not found")
	ErrInvalidClientStatus       = errors.New("client must be on waiting list to move to in care")
	ErrAmbulatoryHoursRequired   = errors.New("ambulatory weekly hours required for ambulatory care")
	ErrAmbulatoryHoursNotAllowed = errors.New("ambulatory weekly hours should only be set for ambulatory care")
	ErrClientNotInCare           = errors.New("client must be in care to be discharged")
	ErrDischargeAlreadyStarted   = errors.New("discharge has already been started for this client")
	ErrDischargeNotStarted       = errors.New("discharge must be started before completing")
)
