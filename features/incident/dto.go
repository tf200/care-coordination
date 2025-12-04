package incident

import "time"

type CreateIncidentRequest struct {
	ClientID            string    `json:"clientId" binding:"required"`
	IncidentDate        time.Time `json:"incidentDate" binding:"required datetime=2006-01-02"`
	IncidentTime        time.Time `json:"incidentTime" binding:"required,datetime=15:04"`
	IncidentType        string    `json:"incidentType" binding:"required,oneof=aggression medical_emergency safety_concern unwanted_behavior other"`
	IncidentSeverity    string    `json:"incidentSeverity" binding:"required,oneof=minor moderate severe"`
	LocationID          string    `json:"locationId" binding:"required"`
	CoordinatorID       string    `json:"coordinatorId" binding:"required"`
	IncidentDescription string    `json:"incidentDescription" binding:"required"`
	ActionTaken         string    `json:"actionTaken" binding:"required"`
	OtherParties        string    `json:"otherParties"`
	Status              string    `json:"status" binding:"required,oneof=pending under_investigation completed"`
}

type CreateIncidentResponse struct {
	ID string `json:"id"`
}

type ListIncidentsRequest struct {
	Search *string `form:"search"`
}

type ListIncidentsResponse struct {
	ID                   string    `json:"id"`
	ClientID             string    `json:"clientId"`
	ClientFirstName      string    `json:"clientFirstName"`
	ClientLastName       string    `json:"clientLastName"`
	IncidentDate         time.Time `json:"incidentDate"`
	IncidentTime         string    `json:"incidentTime"`
	IncidentType         string    `json:"incidentType"`
	IncidentSeverity     string    `json:"incidentSeverity"`
	LocationID           string    `json:"locationId"`
	LocationName         string    `json:"locationName"`
	CoordinatorID        string    `json:"coordinatorId"`
	CoordinatorFirstName string    `json:"coordinatorFirstName"`
	CoordinatorLastName  string    `json:"coordinatorLastName"`
	IncidentDescription  string    `json:"incidentDescription"`
	ActionTaken          string    `json:"actionTaken"`
	OtherParties         *string   `json:"otherParties"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `json:"createdAt"`
}
