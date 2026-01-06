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
	StartTime      time.Time         `json:"start_time" binding:"required"`
	EndTime        time.Time         `json:"end_time" binding:"required,gtfield=StartTime"`
	Location       string            `json:"location"`
	Status         AppointmentStatus `json:"status" binding:"omitempty,oneof=confirmed cancelled tentative"`
	Type           AppointmentType   `json:"type" binding:"required,oneof=general intake ambulatory"`
	RecurrenceRule string            `json:"recurrence_rule"`
	Participants   []ParticipantDTO  `json:"participants" binding:"required,min=1"`
}

type UpdateAppointmentRequest struct {
	Title          *string            `json:"title"`
	Description    *string            `json:"description"`
	StartTime      *time.Time         `json:"start_time"`
	EndTime        *time.Time         `json:"end_time"`
	Location       *string            `json:"location"`
	Status         *AppointmentStatus `json:"status" binding:"omitempty,oneof=confirmed cancelled tentative"`
	Type           *AppointmentType   `json:"type" binding:"omitempty,oneof=general intake ambulatory"`
	RecurrenceRule *string            `json:"recurrence_rule"`
	Participants   []ParticipantDTO   `json:"participants"`
}

type AppointmentResponse struct {
	ID             string            `json:"id"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	StartTime      time.Time         `json:"start_time"`
	EndTime        time.Time         `json:"end_time"`
	Location       string            `json:"location"`
	OrganizerID    string            `json:"organizer_id"`
	Status         AppointmentStatus `json:"status"`
	Type           AppointmentType   `json:"type"`
	RecurrenceRule string            `json:"recurrence_rule"`
	Participants   []ParticipantDTO  `json:"participants"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type CreateReminderRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	DueTime     time.Time `json:"due_time" binding:"required"`
}

type ReminderResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueTime     time.Time `json:"due_time"`
	IsCompleted bool      `json:"is_completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CalendarEventDTO struct {
	ID              string                `json:"id"`
	Title           string                `json:"title"`
	Start           time.Time             `json:"start"`
	End             *time.Time            `json:"end,omitempty"`
	AllDay          bool                  `json:"all_day"`
	Type            string                `json:"type"` // "appointment" or "reminder"
	BackgroundColor string                `json:"background_color,omitempty"`
	ExtendedProps   CalendarExtendedProps `json:"extended_props"`
}

type CalendarExtendedProps struct {
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	Status      string `json:"status,omitempty"`
	Type        string `json:"type,omitempty"`
	IsCompleted bool   `json:"is_completed,omitempty"`
}
