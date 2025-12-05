package intake

import (
	"time"
)

type CreateIntakeFormRequest struct {
	RegistrationFormID string  `json:"registrationFormId" binding:"required"`
	IntakeDate         string  `json:"intakeDate" binding:"required,datetime=2006-01-02"`
	IntakeTime         string  `json:"intakeTime" binding:"required,datetime=15:04:05"`
	LocationID         string  `json:"locationId" binding:"required"`
	CoordinatorID      string  `json:"coordinatorId" binding:"required"`
	FamilySituation    *string `json:"familySituation"`
	MainProvider       *string `json:"mainProvider"`
	Limitations        *string `json:"limitations"`
	FocusAreas         *string `json:"focusAreas"`
	Goals              *string `json:"goals"`
	Notes              *string `json:"notes"`
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
	OrganizationName     *string   `json:"organizationName"`
	LocationName         *string   `json:"locationName"`
	CoordinatorFirstName *string   `json:"coordinatorFirstName"`
	CoordinatorLastName  *string   `json:"coordinatorLastName"`
	Status               string    `json:"status"`
}
