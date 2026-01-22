package calendar

import (
	"context"
	"time"
)

type CalendarService interface {
	// Appointment methods
	CreateAppointment(ctx context.Context, organizerID string, req CreateAppointmentRequest) (*AppointmentResponse, error)
	GetAppointment(ctx context.Context, id string) (*AppointmentResponse, error)
	UpdateAppointment(ctx context.Context, id string, req UpdateAppointmentRequest) (*AppointmentResponse, error)
	DeleteAppointment(ctx context.Context, id string) error
	ListAppointments(ctx context.Context, userID string) ([]AppointmentResponse, error)

	// Reminder methods
	CreateReminder(ctx context.Context, userID string, req CreateReminderRequest) (*ReminderResponse, error)
	GetReminder(ctx context.Context, id string) (*ReminderResponse, error)
	UpdateReminder(ctx context.Context, id string, completed bool) (*ReminderResponse, error)
	DeleteReminder(ctx context.Context, id string) error
	ListReminders(ctx context.Context, userID string) ([]ReminderResponse, error)

	GetCalendarView(ctx context.Context, userID string, startTime, endTime time.Time) ([]CalendarEvent, error)
}
