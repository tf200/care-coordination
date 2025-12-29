package evaluation

import "time"

type GoalProgressDTO struct {
	GoalID        string  `json:"goalId"        binding:"required"`
	Status        string  `json:"status"        binding:"required,oneof=starting on_track delayed achieved"`
	ProgressNotes *string `json:"progressNotes"`
}

type CreateEvaluationRequest struct {
	ClientID       string            `json:"clientId"      binding:"required"`
	CoordinatorID  string            `json:"coordinatorId" binding:"required"`
	EvaluationDate string            `json:"evaluationDate" binding:"required,datetime=2006-01-02"`
	OverallNotes   *string           `json:"overallNotes"`
	ProgressLogs   []GoalProgressDTO `json:"progressLogs"   binding:"min=1"`
	IsDraft        bool              `json:"isDraft"`
}

type CreateEvaluationResponse struct {
	ID                 string     `json:"id"`
	NextEvaluationDate *time.Time `json:"nextEvaluationDate,omitempty"` // Only set when not draft
	IsDraft            bool       `json:"isDraft"`
}

type SaveDraftRequest struct {
	ClientID       string            `json:"clientId"      binding:"required"`
	CoordinatorID  string            `json:"coordinatorId" binding:"required"`
	EvaluationDate string            `json:"evaluationDate" binding:"required,datetime=2006-01-02"`
	OverallNotes   *string           `json:"overallNotes"`
	ProgressLogs   []GoalProgressDTO `json:"progressLogs"`
}

type SaveDraftResponse struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DraftEvaluationListItemDTO struct {
	EvaluationID    string    `json:"evaluationId"`
	ClientID        string    `json:"clientId"`
	ClientFirstName string    `json:"clientFirstName"`
	ClientLastName  string    `json:"clientLastName"`
	EvaluationDate  time.Time `json:"evaluationDate"`
	GoalsCount      int       `json:"goalsCount"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type DraftEvaluationDTO struct {
	EvaluationID         string                `json:"evaluationId"`
	ClientID             string                `json:"clientId"`
	ClientFirstName      string                `json:"clientFirstName"`
	ClientLastName       string                `json:"clientLastName"`
	EvaluationDate       time.Time             `json:"evaluationDate"`
	OverallNotes         *string               `json:"overallNotes"`
	CoordinatorFirstName string                `json:"coordinatorFirstName"`
	CoordinatorLastName  string                `json:"coordinatorLastName"`
	GoalProgress         []GoalProgressItemDTO `json:"goalProgress"`
	CreatedAt            time.Time             `json:"createdAt"`
	UpdatedAt            time.Time             `json:"updatedAt"`
}

type EvaluationHistoryItem struct {
	EvaluationID         string    `json:"evaluationId"`
	EvaluationDate       time.Time `json:"evaluationDate"`
	OverallNotes         *string   `json:"overallNotes"`
	GoalID               string    `json:"goalId"`
	GoalTitle            string    `json:"goalTitle"`
	Status               string    `json:"status"`
	ProgressNotes        *string   `json:"progressNotes"`
	CoordinatorFirstName string    `json:"coordinatorFirstName"`
	CoordinatorLastName  string    `json:"coordinatorLastName"`
}

type UpcomingEvaluationDTO struct {
	ID                      string    `json:"id"`
	FirstName               string    `json:"firstName"`
	LastName                string    `json:"lastName"`
	NextEvaluationDate      time.Time `json:"nextEvaluationDate"`
	EvaluationIntervalWeeks int       `json:"evaluationIntervalWeeks"`
	LocationName            string    `json:"locationName"`
	CoordinatorFirstName    string    `json:"coordinatorFirstName"`
	CoordinatorLastName     string    `json:"coordinatorLastName"`
	HasDraft                bool      `json:"hasDraft"`
}

type GlobalRecentEvaluationDTO struct {
	EvaluationID         string    `json:"evaluationId"`
	ClientID             string    `json:"clientId"`
	EvaluationDate       time.Time `json:"evaluationDate"`
	ClientFirstName      string    `json:"clientFirstName"`
	ClientLastName       string    `json:"clientLastName"`
	CoordinatorFirstName string    `json:"coordinatorFirstName"`
	CoordinatorLastName  string    `json:"coordinatorLastName"`
	TotalGoals           int       `json:"totalGoals"`
	GoalsAchieved        int       `json:"goalsAchieved"`
}

type GoalProgressItemDTO struct {
	GoalID        string  `json:"goalId"`
	GoalTitle     string  `json:"goalTitle"`
	Status        string  `json:"status"`
	ProgressNotes *string `json:"progressNotes"`
}

type CoordinatorInfoDTO struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type LastEvaluationDTO struct {
	EvaluationID   string                `json:"evaluationId"`
	EvaluationDate time.Time             `json:"evaluationDate"`
	OverallNotes   *string               `json:"overallNotes"`
	Coordinator    CoordinatorInfoDTO    `json:"coordinator"`
	GoalProgress   []GoalProgressItemDTO `json:"goalProgress"`
}
