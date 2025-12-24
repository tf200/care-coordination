package client

type MoveClientToWaitingListRequest struct {
	IntakeFormID        string `json:"intakeFormId"`
	WaitingListPriority string `json:"waitingListPriority" binding:"required,oneof=low normal high"`
}

type MoveClientToWaitingListResponse struct {
	ClientID string `json:"clientId"`
}

type MoveClientInCareRequest struct {
	CareStartDate         string `json:"careStartDate"         binding:"required format=2006-01-02"`
	CareEndDate           string `json:"careEndDate"           binding:"required format=2006-01-02"`
	AmbulatoryWeeklyHours *int32 `json:"ambulatoryWeeklyHours"`
}

type MoveClientInCareResponse struct {
	ClientID string `json:"clientId"`
}

// Phase 1: Start Discharge - initiates discharge process, client remains in_care
type StartDischargeRequest struct {
	DischargeDate      string `json:"dischargeDate"      binding:"required,datetime=2006-01-02"`
	ReasonForDischarge string `json:"reasonForDischarge" binding:"required,oneof=treatment_completed terminated_by_mutual_agreement terminated_by_client terminated_by_provider terminated_due_to_external_factors other"`
}

type StartDischargeResponse struct {
	ClientID string `json:"clientId"`
}

// Phase 2: Complete Discharge - finalizes discharge, requires reports
type CompleteDischargeRequest struct {
	ClosingReport          string   `json:"closingReport"          binding:"required"`
	EvaluationReport       string   `json:"evaluationReport"       binding:"required"`
	DischargeAttachmentIDs []string `json:"dischargeAttachmentIds"`
}

type CompleteDischargeResponse struct {
	ClientID string `json:"clientId"`
}

type ListWaitingListClientsRequest struct {
	Search *string `form:"search"`
}

type ListWaitingListClientsResponse struct {
	ID                   string  `json:"id"`
	FirstName            string  `json:"firstName"`
	LastName             string  `json:"lastName"`
	Bsn                  string  `json:"bsn"`
	DateOfBirth          string  `json:"dateOfBirth"`
	PhoneNumber          *string `json:"phoneNumber"`
	Gender               string  `json:"gender"`
	CareType             string  `json:"careType"`
	WaitingListPriority  string  `json:"waitingListPriority"`
	FocusAreas           *string `json:"focusAreas"`
	Notes                *string `json:"notes"`
	CreatedAt            string  `json:"createdAt"`
	LocationID           string  `json:"locationId"`
	LocationName         string  `json:"locationName"`
	CoordinatorID        string  `json:"coordinatorId"`
	CoordinatorFirstName string  `json:"coordinatorFirstName"`
	CoordinatorLastName  string  `json:"coordinatorLastName"`
	ReferringOrgName     *string `json:"referringOrgName"`
}

type ListInCareClientsRequest struct {
	Search *string `form:"search"`
}

type ListInCareClientsResponse struct {
	ID                   string  `json:"id"`
	FirstName            string  `json:"firstName"`
	LastName             string  `json:"lastName"`
	Bsn                  string  `json:"bsn"`
	DateOfBirth          string  `json:"dateOfBirth"`
	PhoneNumber          *string `json:"phoneNumber"`
	Gender               string  `json:"gender"`
	CareType             string  `json:"careType"`
	CareStartDate        string  `json:"careStartDate"`
	CareEndDate          string  `json:"careEndDate"`
	LocationID           string  `json:"locationId"`
	LocationName         string  `json:"locationName"`
	CoordinatorID        string  `json:"coordinatorId"`
	CoordinatorFirstName string  `json:"coordinatorFirstName"`
	CoordinatorLastName  string  `json:"coordinatorLastName"`
	ReferringOrgName     *string `json:"referringOrgName"`
	// For protected_living, semi_independent_living, independent_assisted_living
	WeeksInAccommodation *int `json:"weeksInAccommodation,omitempty"`
	// For ambulatory_care only
	UsedAmbulatoryHours *int `json:"usedAmbulatoryHours,omitempty"`
}

type ListDischargedClientsRequest struct {
	Search          *string `form:"search"`
	DischargeStatus *string `form:"dischargeStatus" binding:"omitempty,oneof=in_progress completed"`
}

type ListDischargedClientsResponse struct {
	ID                   string  `json:"id"`
	FirstName            string  `json:"firstName"`
	LastName             string  `json:"lastName"`
	Bsn                  string  `json:"bsn"`
	DateOfBirth          string  `json:"dateOfBirth"`
	PhoneNumber          *string `json:"phoneNumber"`
	Gender               string  `json:"gender"`
	CareType             string  `json:"careType"`
	CareStartDate        string  `json:"careStartDate"`
	DischargeDate        string  `json:"dischargeDate"`
	ReasonForDischarge   string  `json:"reasonForDischarge"`
	DischargeStatus      string  `json:"dischargeStatus"`
	ClosingReport        *string `json:"closingReport"`
	EvaluationReport     *string `json:"evaluationReport"`
	LocationID           string  `json:"locationId"`
	LocationName         string  `json:"locationName"`
	CoordinatorID        string  `json:"coordinatorId"`
	CoordinatorFirstName string  `json:"coordinatorFirstName"`
	CoordinatorLastName  string  `json:"coordinatorLastName"`
	ReferringOrgName     *string `json:"referringOrgName"`
}

type PriorityCountsDTO struct {
	Low    int `json:"low"`
	Normal int `json:"normal"`
	High   int `json:"high"`
}

type GetWaitlistStatsResponse struct {
	TotalCount         int               `json:"totalCount"`
	AverageDaysWaiting float64           `json:"averageDaysWaiting"`
	HighPriorityCount  int               `json:"highPriorityCount"`
	CountsByPriority   PriorityCountsDTO `json:"countsByPriority"`
}

type CareTypeCountsDTO struct {
	ProtectedLiving           int `json:"protectedLiving"`
	SemiIndependentLiving     int `json:"semiIndependentLiving"`
	IndependentAssistedLiving int `json:"independentAssistedLiving"`
	AmbulatoryCare            int `json:"ambulatoryCare"`
}

type GetInCareStatsResponse struct {
	TotalCount        int               `json:"totalCount"`
	AverageDaysInCare float64           `json:"averageDaysInCare"`
	CountsByCareType  CareTypeCountsDTO `json:"countsByCareType"`
}

type GetDischargeStatsResponse struct {
	TotalCount              int     `json:"totalCount"`
	CompletedDischarges     int     `json:"completedDischarges"`
	PrematureDischarges     int     `json:"prematureDischarges"`
	DischargeCompletionRate float64 `json:"dischargeCompletionRate"`
	AverageDaysInCare       float64 `json:"averageDaysInCare"`
}
