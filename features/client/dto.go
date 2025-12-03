package client

import "time"

type MoveClientToWaitingListRequest struct {
	IntakeFormID        string `json:"intakeFormId"`
	WaitingListPriority string `json:"waitingListPriority" binding:"required,oneof=low normal high"`
}

type MoveClientToWaitingListResponse struct {
	ClientID string `json:"clientId"`
}

type MoveClientInCareRequest struct {
	CareStartDate time.Time `json:"careStartDate" binding:"required format=2006-01-02"`
	CareEndDate   time.Time `json:"careEndDate" binding:"omitempty format=2006-01-02"`
}

type MoveClientInCareResponse struct {
	ClientID string `json:"clientId"`
}
