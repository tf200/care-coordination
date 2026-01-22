package intake

import (
	"time"
)

type GoalItem struct {
	ID          *string `json:"id"`
	Title       string  `json:"title"       binding:"required"`
	Description *string `json:"description"`
}

type CreateIntakeFormRequest struct {
	RegistrationFormID string     `json:"registrationFormId" binding:"required"`
	IntakeDate         string     `json:"intakeDate"         binding:"required,datetime=2006-01-02"`
	IntakeTime         string     `json:"intakeTime"         binding:"required,datetime=15:04"`
	LocationID         string     `json:"locationId"         binding:"required"`
	CoordinatorID      string     `json:"coordinatorId"      binding:"required"`
	FamilySituation    *string    `json:"familySituation"`
	MainProvider       *string    `json:"mainProvider"`
	Limitations        *string    `json:"limitations"`
	FocusAreas         *string    `json:"focusAreas"`
	Goals              []GoalItem `json:"goals"              binding:"min=1"`
	Notes              *string    `json:"notes"`
	EvaluationInterval *int       `json:"evaluationIntervalWeeks" binding:"omitempty,min=1"`
}

type CreateIntakeFormResponse struct {
	ID string `json:"id"`
}

type ListIntakeFormsRequest struct {
	Search *string `form:"search"`
}

type ListIntakeFormsResponse struct {
	ID                   string    `json:"id"`
	RegistrationFormID   string    `json:"registrationFormId"`
	IntakeDate           time.Time `json:"intakeDate"`
	IntakeTime           string    `json:"intakeTime"`
	LocationID           string    `json:"locationId"`
	CoordinatorID        string    `json:"coordinatorId"`
	MainProvider         *string   `json:"mainProvider"`
	UpdatedAt            time.Time `json:"updatedAt"`
	ClientFirstName      *string   `json:"clientFirstName"`
	ClientLastName       *string   `json:"clientLastName"`
	ClientBSN            *string   `json:"clientBsn"`
	CareType             *string   `json:"careType"`
	OrganizationName     *string   `json:"organizationName"`
	LocationName         *string   `json:"locationName"`
	CoordinatorFirstName *string   `json:"coordinatorFirstName"`
	CoordinatorLastName  *string   `json:"coordinatorLastName"`
	Status               string    `json:"status"`
}

type GetIntakeFormResponse struct {
	ID                   string     `json:"id"`
	RegistrationFormID   string     `json:"registrationFormId"`
	IntakeDate           time.Time  `json:"intakeDate"`
	IntakeTime           string     `json:"intakeTime"`
	LocationID           string     `json:"locationId"`
	CoordinatorID        string     `json:"coordinatorId"`
	FamilySituation      *string    `json:"familySituation"`
	MainProvider         *string    `json:"mainProvider"`
	Limitations          *string    `json:"limitations"`
	FocusAreas           *string    `json:"focusAreas"`
	Goals                []GoalItem `json:"goals"`
	Notes                *string    `json:"notes"`
	EvaluationInterval   int        `json:"evaluationIntervalWeeks"`
	Status               string     `json:"status"`
	ClientFirstName      *string    `json:"clientFirstName"`
	ClientLastName       *string    `json:"clientLastName"`
	ClientBSN            *string    `json:"clientBsn"`
	CareType             *string    `json:"careType"`
	OrganizationName     *string    `json:"organizationName"`
	LocationName         *string    `json:"locationName"`
	CoordinatorFirstName *string    `json:"coordinatorFirstName"`
	CoordinatorLastName  *string    `json:"coordinatorLastName"`
	HasClient            bool       `json:"hasClient"`
}

type UpdateIntakeFormRequest struct {
	IntakeDate         *string    `json:"intakeDate"      binding:"omitempty,datetime=2006-01-02"`
	IntakeTime         *string    `json:"intakeTime"      binding:"omitempty,datetime=15:04"`
	LocationID         *string    `json:"locationId"`
	CoordinatorID      *string    `json:"coordinatorId"`
	FamilySituation    *string    `json:"familySituation"`
	MainProvider       *string    `json:"mainProvider"`
	Limitations        *string    `json:"limitations"`
	FocusAreas         *string    `json:"focusAreas"`
	Goals              []GoalItem `json:"goals"`
	Notes              *string    `json:"notes"`
	EvaluationInterval *int       `json:"evaluationIntervalWeeks" binding:"omitempty,min=1"`
	Status             *string    `json:"status"          binding:"omitempty,oneof=completed pending"`
}

type UpdateIntakeFormResponse struct {
	ID string `json:"id"`
}

type GetIntakeStatsResponse struct {
	TotalCount           int     `json:"totalCount"`
	PendingCount         int     `json:"pendingCount"`
	ConversionPercentage float64 `json:"conversionPercentage"`
}
