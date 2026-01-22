package dashboard

type OverviewResponse struct {
	TotalActiveClients   int `json:"totalActiveClients"`
	WaitingListCount     int `json:"waitingListCount"`
	PendingRegistrations int `json:"pendingRegistrations"`
	TotalCoordinators    int `json:"totalCoordinators"`
	TotalEmployees       int `json:"totalEmployees"`
	OpenIncidents        int `json:"openIncidents"`
}

// Alert severity levels
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
)

// Alert types
type AlertType string

const (
	AlertTypeEvaluation AlertType = "evaluation"
	AlertTypeCareEnd    AlertType = "care_end"
	AlertTypeIncident   AlertType = "incident"
	AlertTypeWaitlist   AlertType = "waitlist"
	AlertTypeTransfer   AlertType = "transfer"
)

type AlertItem struct {
	ID          string        `json:"id"`
	Type        AlertType     `json:"type"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Severity    AlertSeverity `json:"severity"`
	Count       int           `json:"count"`
	Link        string        `json:"link"`
}

type CriticalAlertsResponse struct {
	Alerts []AlertItem `json:"alerts"`
}

type PipelineStatsResponse struct {
	Registrations int `json:"registrations"`
	Intakes       int `json:"intakes"`
	WaitingList   int `json:"waitingList"`
	InCare        int `json:"inCare"`
	Discharged    int `json:"discharged"`
}

type CareTypeDistributionItem struct {
	CareType   string  `json:"careType"`
	Label      string  `json:"label"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type CareTypeDistributionResponse struct {
	Distribution []CareTypeDistributionItem `json:"distribution"`
	Total        int                        `json:"total"`
}

type LocationCapacityRequest struct {
	Limit int    `form:"limit,default=4" binding:"min=1,max=100"`
	Sort  string `form:"sort,default=occupancy_desc" binding:"omitempty,oneof=occupancy_desc occupancy_asc name"`
}

type LocationCapacityItem struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Capacity   int     `json:"capacity"`
	Occupied   int     `json:"occupied"`
	Available  int     `json:"available"`
	Percentage float64 `json:"percentage"`
}

type LocationCapacityTotals struct {
	TotalCapacity     int     `json:"totalCapacity"`
	TotalOccupied     int     `json:"totalOccupied"`
	TotalAvailable    int     `json:"totalAvailable"`
	OverallPercentage float64 `json:"overallPercentage"`
}

type LocationCapacityResponse struct {
	Locations []LocationCapacityItem `json:"locations"`
	Totals    LocationCapacityTotals `json:"totals"`
}

type TodayAppointmentItem struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	ClientID     string `json:"clientId,omitempty"`
	ClientName   string `json:"clientName,omitempty"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	LocationName string `json:"locationName,omitempty"`
}

type TodayAppointmentsResponse struct {
	Appointments []TodayAppointmentItem `json:"appointments"`
	Count        int                    `json:"count"`
}

type EvaluationStatsResponse struct {
	CompletionRate int `json:"completionRate"`
	Completed      int `json:"completed"`
	Total          int `json:"total"`
	Overdue        int `json:"overdue"`
	DueSoon        int `json:"dueSoon"`
}

type DischargeStatsResponse struct {
	ThisMonth         int `json:"thisMonth"`
	ThisYear          int `json:"thisYear"`
	PlannedRate       int `json:"plannedRate"`
	AverageDaysInCare int `json:"averageDaysInCare"`
}

// Coordinator Dashboard DTOs

type CoordinatorAlertSeverity string

const (
	CoordinatorAlertSeverityCritical CoordinatorAlertSeverity = "critical"
	CoordinatorAlertSeverityWarning  CoordinatorAlertSeverity = "warning"
	CoordinatorAlertSeverityInfo     CoordinatorAlertSeverity = "info"
)

type CoordinatorAlertType string

const (
	CoordinatorAlertTypeEvaluation  CoordinatorAlertType = "evaluation"
	CoordinatorAlertTypeContract    CoordinatorAlertType = "contract"
	CoordinatorAlertTypeDraft       CoordinatorAlertType = "draft"
	CoordinatorAlertTypeIncident    CoordinatorAlertType = "incident"
	CoordinatorAlertTypeWaitingLong CoordinatorAlertType = "waiting_long"
)

type CoordinatorUrgentAlertItem struct {
	ID          string                   `json:"id"`
	Type        CoordinatorAlertType     `json:"type"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Severity    CoordinatorAlertSeverity `json:"severity"`
	Count       int                      `json:"count"`
	ClientIDs   []string                 `json:"clientIds"`
	Link        string                   `json:"link"`
}

type CoordinatorUrgentAlertsResponse struct {
	Alerts []CoordinatorUrgentAlertItem `json:"alerts"`
}

type CoordinatorScheduleItem struct {
	ID           string `json:"id"`
	Time         string `json:"time"`
	EndTime      string `json:"endTime"`
	Type         string `json:"type"`
	ClientID     string `json:"clientId,omitempty"`
	ClientName   string `json:"clientName"`
	LocationID   string `json:"locationId,omitempty"`
	LocationName string `json:"locationName"`
	Status       string `json:"status"`
}

type CoordinatorTodayScheduleResponse struct {
	Date         string                    `json:"date"`
	Appointments []CoordinatorScheduleItem `json:"appointments"`
	Count        int                       `json:"count"`
}

type CoordinatorStatsResponse struct {
	MyActiveClients       int `json:"myActiveClients"`
	MyUpcomingEvaluations int `json:"myUpcomingEvaluations"`
	MyPendingIntakes      int `json:"myPendingIntakes"`
	MyWaitingListClients  int `json:"myWaitingListClients"`
}

type ReminderItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Client   string `json:"client"`
	DueDate  string `json:"dueDate"`
	Priority string `json:"priority"`
}

type CoordinatorRemindersResponse struct {
	Reminders []ReminderItem `json:"reminders"`
}

type CoordinatorClientItem struct {
	ID               string `json:"id"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	CareType         string `json:"careType"`
	Location         string `json:"location"`
	DaysUntilEnd     int    `json:"daysUntilEnd"`
	Status           string `json:"status"`
	NextEvaluation   string `json:"nextEvaluation"`
	EvaluationStatus string `json:"evaluationStatus"`
}

type CoordinatorClientsResponse struct {
	Clients []CoordinatorClientItem `json:"clients"`
}

type CoordinatorGoalsProgressResponse struct {
	OnTrack    int `json:"onTrack"`
	Delayed    int `json:"delayed"`
	Achieved   int `json:"achieved"`
	NotStarted int `json:"notStarted"`
	Total      int `json:"total"`
}

type CoordinatorIncidentItem struct {
	ID       string `json:"id"`
	Client   string `json:"client"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Date     string `json:"date"`
	Status   string `json:"status"`
}

type CoordinatorIncidentsResponse struct {
	Incidents []CoordinatorIncidentItem `json:"incidents"`
}
