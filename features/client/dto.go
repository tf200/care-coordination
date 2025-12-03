package client

import "github.com/jackc/pgx/v5/pgtype"

type MoveClientToWaitingListRequest struct {
	IntakeFormID        string `json:"intakeFormId"`
	WaitingListPriority string `json:"waitingListPriority" binding:"required,oneof=low normal high"`
}

type MoveClientToWaitingListResponse struct {
	ClientID string `json:"clientId"`
}

type MoveClientInCareRequest struct {
	CareStartDate         pgtype.Date `json:"careStartDate" binding:"required"`
	CareEndDate           pgtype.Date `json:"careEndDate"`
	AmbulatoryWeeklyHours *int32      `json:"ambulatoryWeeklyHours"`
}

type MoveClientInCareResponse struct {
	ClientID string `json:"clientId"`
}
