package incident

import "time"

type CreateIncidentRequest struct {
	ClientID            string `json:"clientId"       binding:"required"`
	IncidentDate        string `json:"incidentDate"   binding:"required,datetime=2006-01-02"`
	IncidentTime        string `json:"incidentTime"   binding:"required,datetime=15:04"`
	IncidentType        string `json:"incidentType"        binding:"required,oneof=aggression medical_emergency safety_concern unwanted_behavior other"`
	IncidentSeverity    string `json:"incidentSeverity"    binding:"required,oneof=minor moderate severe"`
	LocationID          string `json:"locationId"          binding:"required"`
	CoordinatorID       string `json:"coordinatorId"       binding:"required"`
	IncidentDescription string `json:"incidentDescription" binding:"required"`
	ActionTaken         string `json:"actionTaken"         binding:"required"`
	OtherParties        string `json:"otherParties"`
	Status              string `json:"status"              binding:"required,oneof=pending under_investigation completed"`
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

type IncidentSeverityCountsDTO struct {
	Minor    int `json:"minor"`
	Moderate int `json:"moderate"`
	Severe   int `json:"severe"`
}

type IncidentStatusCountsDTO struct {
	Pending            int `json:"pending"`
	UnderInvestigation int `json:"underInvestigation"`
	Completed          int `json:"completed"`
}

type IncidentTypeCountsDTO struct {
	Aggression       int `json:"aggression"`
	MedicalEmergency int `json:"medicalEmergency"`
	SafetyConcern    int `json:"safetyConcern"`
	UnwantedBehavior int `json:"unwantedBehavior"`
	Other            int `json:"other"`
}

type GetIncidentStatsResponse struct {
	TotalCount       int                       `json:"totalCount"`
	CountsBySeverity IncidentSeverityCountsDTO `json:"countsBySeverity"`
	CountsByStatus   IncidentStatusCountsDTO   `json:"countsByStatus"`
	CountsByType     IncidentTypeCountsDTO     `json:"countsByType"`
}

type GetIncidentResponse struct {
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

type UpdateIncidentRequest struct {
	IncidentDate        *string `json:"incidentDate"   binding:"omitempty,datetime=2006-01-02"`
	IncidentTime        *string `json:"incidentTime"   binding:"omitempty,datetime=15:04"`
	IncidentType        *string `json:"incidentType"        binding:"omitempty,oneof=aggression medical_emergency safety_concern unwanted_behavior other"`
	IncidentSeverity    *string `json:"incidentSeverity"    binding:"omitempty,oneof=minor moderate severe"`
	LocationID          *string `json:"locationId"`
	CoordinatorID       *string `json:"coordinatorId"`
	IncidentDescription *string `json:"incidentDescription"`
	ActionTaken         *string `json:"actionTaken"`
	OtherParties        *string `json:"otherParties"`
	Status              *string `json:"status"              binding:"omitempty,oneof=pending under_investigation completed"`
}

type UpdateIncidentResponse struct {
	Success bool `json:"success"`
}

type DeleteIncidentResponse struct {
	Success bool `json:"success"`
}
