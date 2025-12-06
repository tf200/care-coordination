package client

type MoveClientToWaitingListRequest struct {
	IntakeFormID        string `json:"intakeFormId"`
	WaitingListPriority string `json:"waitingListPriority" binding:"required,oneof=low normal high"`
}

type MoveClientToWaitingListResponse struct {
	ClientID string `json:"clientId"`
}

type MoveClientInCareRequest struct {
	CareStartDate         string `json:"careStartDate" binding:"required format=2006-01-02"`
	CareEndDate           string `json:"careEndDate" binding:"required format=2006-01-02"`
	AmbulatoryWeeklyHours *int32 `json:"ambulatoryWeeklyHours"`
}

type MoveClientInCareResponse struct {
	ClientID string `json:"clientId"`
}

type DischargeClientRequest struct {
	DischargeDate          string   `json:"dischargeDate" binding:"required format=2006-01-02"`
	ClosingReport          *string  `json:"closingReport"`
	EvaluationReport       *string  `json:"evaluationReport"`
	ReasonForDischarge     string   `json:"reasonForDischarge" binding:"required,oneof=treatment_completed terminated_by_mutual_agreement terminated_by_client terminated_by_provider terminated_due_to_external_factors other"`
	DischargeAttachmentIDs []string `json:"dischargeAttachmentIds"`
	DischargeStatus        string   `json:"dischargeStatus" binding:"required,oneof=in_progress completed"`
}

type DischargeClientResponse struct {
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
