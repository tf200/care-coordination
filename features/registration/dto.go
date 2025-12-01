package registration

import (
	"time"
)

type CreateRegistrationFormRequest struct {
	FirstName          string    `json:"firstName" binding:"required"`
	LastName           string    `json:"lastName" binding:"required"`
	BSN                string    `json:"bsn" binding:"required"`
	DateOfBirth        time.Time `json:"dateOfBirth" binding:"required format=2006-01-02"`
	OrgName            string    `json:"orgName" binding:"required"`
	OrgContactPerson   string    `json:"orgContactPerson" binding:"required"`
	OrgPhoneNumber     string    `json:"orgPhoneNumber" binding:"required"`
	OrgEmail           string    `json:"orgEmail" binding:"required"`
	CareType           string    `json:"careType" binding:"required"`
	CoordinatorID      string    `json:"coordinatorId" binding:"required"`
	RegistrationDate   time.Time `json:"registrationDate" binding:"required format=2006-01-02"`
	RegistrationReason string    `json:"registrationReason" binding:"required"`
	AdditionalNotes    *string   `json:"additionalNotes"`
	AttachmentIDs      []string  `json:"attachmentIds"`
}

type CreateRegistrationFormResponse struct {
	ID string `json:"id"`
}

type ListRegistrationFormsResponse struct {
	ID                   string    `json:"id"`
	FirstName            string    `json:"firstName"`
	LastName             string    `json:"lastName"`
	Bsn                  string    `json:"bsn"`
	DateOfBirth          time.Time `json:"dateOfBirth"`
	OrgName              string    `json:"orgName"`
	CareType             string    `json:"careType"`
	CoordinatorID        string    `json:"coordinatorId"`
	CoordinatorFirstName string    `json:"coordinatorFirstName"`
	CoordinatorLastName  string    `json:"coordinatorLastName"`
	RegistrationDate     time.Time `json:"registrationDate"`
	RegistrationReason   string    `json:"registrationReason"`
	NumberOfAttachments  int       `json:"numberOfAttachments"`
}
