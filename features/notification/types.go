package notification

// Notification type constants matching the database enum
const (
	TypeEvaluationDue            = "evaluation_due"
	TypeAppointmentReminder      = "appointment_reminder"
	TypeIncidentCreated          = "incident_created"
	TypeLocationTransferRequest  = "location_transfer_request"
	TypeLocationTransferApproved = "location_transfer_approved"
	TypeLocationTransferRejected = "location_transfer_rejected"
	TypeClientStatusChange       = "client_status_change"
	TypeRegistrationStatusChange = "registration_status_change"
	TypeSystemAlert              = "system_alert"
)

// Notification priority constants matching the database enum
const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// Resource type constants for linking notifications to entities
const (
	ResourceTypeClient           = "client"
	ResourceTypeIncident         = "incident"
	ResourceTypeAppointment      = "appointment"
	ResourceTypeEvaluation       = "evaluation"
	ResourceTypeLocationTransfer = "location_transfer"
	ResourceTypeRegistration     = "registration"
)
