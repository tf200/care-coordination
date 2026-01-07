package calendar

import (
	"time"
)

type AppointmentStatus string

const (
	StatusConfirmed AppointmentStatus = "confirmed"
	StatusCancelled AppointmentStatus = "cancelled"
	StatusTentative AppointmentStatus = "tentative"
)

type AppointmentType string

const (
	TypeGeneral    AppointmentType = "general"
	TypeIntake     AppointmentType = "intake"
	TypeAmbulatory AppointmentType = "ambulatory"
)

type ParticipantType string

const (
	ParticipantEmployee ParticipantType = "employee"
	ParticipantClient   ParticipantType = "client"
)

type ParticipantDTO struct {
	ID   string          `json:"id" binding:"required"`
	Type ParticipantType `json:"type" binding:"required,oneof=employee client"`
}

type CreateAppointmentRequest struct {
	Title          string            `json:"title" binding:"required"`
	Description    string            `json:"description"`
	StartTime      time.Time         `json:"startTime" binding:"required"`
	EndTime        time.Time         `json:"endTime" binding:"required,gtfield=StartTime"`
	Location       string            `json:"location"`
	Status         AppointmentStatus `json:"status" binding:"omitempty,oneof=confirmed cancelled tentative"`
	Type           AppointmentType   `json:"type" binding:"required,oneof=general intake ambulatory"`
	RecurrenceRule string            `json:"recurrenceRule"`
	Participants   []ParticipantDTO  `json:"participants" binding:"required,min=1"`
}

type UpdateAppointmentRequest struct {
	Title          *string            `json:"title"`
	Description    *string            `json:"description"`
	StartTime      *time.Time         `json:"startTime"`
	EndTime        *time.Time         `json:"endTime"`
	Location       *string            `json:"location"`
	Status         *AppointmentStatus `json:"status" binding:"omitempty,oneof=confirmed cancelled tentative"`
	Type           *AppointmentType   `json:"type" binding:"omitempty,oneof=general intake ambulatory"`
	RecurrenceRule *string            `json:"recurrenceRule"`
	Participants   []ParticipantDTO   `json:"participants"`
}

type AppointmentResponse struct {
	ID             string            `json:"id"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	StartTime      time.Time         `json:"startTime"`
	EndTime        time.Time         `json:"endTime"`
	Location       string            `json:"location"`
	OrganizerID    string            `json:"organizerId"`
	Status         AppointmentStatus `json:"status"`
	Type           AppointmentType   `json:"type"`
	RecurrenceRule string            `json:"recurrenceRule"`
	Participants   []ParticipantDTO  `json:"participants"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

type CreateReminderRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	DueTime     time.Time `json:"dueTime" binding:"required"`
}

type ReminderResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueTime     time.Time `json:"dueTime"`
	IsCompleted bool      `json:"isCompleted"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UpdateReminderRequest struct {
	IsCompleted bool `json:"isCompleted"`
}

type GetCalendarViewRequest struct {
	Start      time.Time `form:"start" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	End        time.Time `form:"end" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	EmployeeID *string   `form:"employeeId"`
}

type CalendarEventDTO struct {
	ID              string                `json:"id"`
	Title           string                `json:"title"`
	Start           time.Time             `json:"start"`
	End             *time.Time            `json:"end,omitempty"`
	AllDay          bool                  `json:"allDay"`
	Type            string                `json:"type"` // "appointment" or "reminder"
	BackgroundColor string                `json:"backgroundColor,omitempty"`
	ExtendedProps   CalendarExtendedProps `json:"extendedProps"`
}

type CalendarExtendedProps struct {
	Description     string `json:"description,omitempty"`
	Location        string `json:"location,omitempty"`
	Status          string `json:"status,omitempty"`
	Type            string `json:"type,omitempty"`
	IsCompleted     bool   `json:"isCompleted,omitempty"`
	IsRecurring     bool   `json:"isRecurring,omitempty"`
	OriginalEventID string `json:"originalEventId,omitempty"`
}
