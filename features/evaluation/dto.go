package evaluation

import "time"

type GoalProgressItem struct {
	GoalID        string  `json:"goalId"`
	GoalTitle     string  `json:"goalTitle"`
	Status        string  `json:"status"`
	ProgressNotes *string `json:"progressNotes"`
}

type CreateEvaluationRequest struct {
	ClientID       string             `json:"clientId"      binding:"required"`
	CoordinatorID  string             `json:"coordinatorId" binding:"required"`
	EvaluationDate string             `json:"evaluationDate" binding:"required,datetime=2006-01-02"`
	OverallNotes   *string            `json:"overallNotes"`
	ProgressLogs   []GoalProgressItem `json:"progressLogs"   binding:"min=1"`
	IsDraft        bool               `json:"isDraft"`
}

type CreateEvaluationResponse struct {
	ID                 string     `json:"id"`
	NextEvaluationDate *time.Time `json:"nextEvaluationDate,omitempty"`
	IsDraft            bool       `json:"isDraft"`
}

type UpdateEvaluationRequest struct {
	EvaluationDate string             `json:"evaluationDate" binding:"required,datetime=2006-01-02"`
	OverallNotes   *string            `json:"overallNotes"`
	ProgressLogs   []GoalProgressItem `json:"progressLogs" binding:"required,min=1"`
}

type UpdateEvaluationResponse struct {
	ID string `json:"id"`
}

type SaveDraftRequest struct {
	ClientID       string             `json:"clientId"      binding:"required"`
	CoordinatorID  string             `json:"coordinatorId" binding:"required"`
	EvaluationDate string             `json:"evaluationDate" binding:"required,datetime=2006-01-02"`
	OverallNotes   *string            `json:"overallNotes"`
	ProgressLogs   []GoalProgressItem `json:"progressLogs"`
}

type SaveDraftResponse struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DraftEvaluationListItem struct {
	EvaluationID    string    `json:"evaluationId"`
	ClientID        string    `json:"clientId"`
	ClientFirstName string    `json:"clientFirstName"`
	ClientLastName  string    `json:"clientLastName"`
	EvaluationDate  time.Time `json:"evaluationDate"`
	GoalsCount      int       `json:"goalsCount"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type DraftEvaluationResponse struct {
	EvaluationID         string             `json:"evaluationId"`
	ClientID             string             `json:"clientId"`
	ClientFirstName      string             `json:"clientFirstName"`
	ClientLastName       string             `json:"clientLastName"`
	EvaluationDate       time.Time          `json:"evaluationDate"`
	OverallNotes         *string            `json:"overallNotes"`
	CoordinatorFirstName string             `json:"coordinatorFirstName"`
	CoordinatorLastName  string             `json:"coordinatorLastName"`
	GoalProgress         []GoalProgressItem `json:"goalProgress"`
	CreatedAt            time.Time          `json:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt"`
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

type UpcomingEvaluationItem struct {
	ID                      string    `json:"id"`
	FirstName               string    `json:"firstName"`
	LastName                string    `json:"lastName"`
	NextEvaluationDate      time.Time `json:"nextEvaluationDate"`
	EvaluationIntervalWeeks int       `json:"evaluationIntervalWeeks"`
	LocationName            string    `json:"locationName"`
	CoordinatorFirstName    string    `json:"coordinatorFirstName"`
	CoordinatorLastName     string    `json:"coordinatorLastName"`
	HasDraft                bool      `json:"hasDraft"`
	DraftID                 *string   `json:"draftId,omitempty"`
}

type GlobalRecentEvaluationItem struct {
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

type CoordinatorInfo struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type LastEvaluationItem struct {
	EvaluationID   string             `json:"evaluationId"`
	EvaluationDate time.Time          `json:"evaluationDate"`
	OverallNotes   *string            `json:"overallNotes"`
	Coordinator    CoordinatorInfo    `json:"coordinator"`
	GoalProgress   []GoalProgressItem `json:"goalProgress"`
}

type EvaluationResponse struct {
	EvaluationID         string             `json:"evaluationId"`
	ClientID             string             `json:"clientId"`
	ClientFirstName      string             `json:"clientFirstName"`
	ClientLastName       string             `json:"clientLastName"`
	EvaluationDate       time.Time          `json:"evaluationDate"`
	OverallNotes         *string            `json:"overallNotes"`
	Status               string             `json:"status"`
	CoordinatorFirstName string             `json:"coordinatorFirstName"`
	CoordinatorLastName  string             `json:"coordinatorLastName"`
	GoalProgress         []GoalProgressItem `json:"goalProgress"`
	CreatedAt            time.Time          `json:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt"`
}
