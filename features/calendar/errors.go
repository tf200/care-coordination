package calendar

import "errors"

var (
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrReminderNotFound    = errors.New("reminder not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInternal            = errors.New("internal server error")
	ErrInvalidRequest      = errors.New("invalid request")
)
