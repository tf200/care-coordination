package registration

import (
	"time"
)

type CreateRegistrationFormRequest struct {
	FirstName          string    `json:"firstName" binding:"required"`
	LastName           string    `json:"lastName" binding:"required"`
	BSN                string    `json:"bsn" binding:"required"`
	DateOfBirth        time.Time `json:"dateOfBirth" binding:"required format=2006-01-02"`
	RefferingOrgID     *string   `json:"refferingOrgId" binding:"required"`
	CareType           string    `json:"careType" binding:"required"`
	RegistrationDate   time.Time `json:"registrationDate" binding:"required format=2006-01-02"`
	RegistrationReason string    `json:"registrationReason" binding:"required"`
	AdditionalNotes    *string   `json:"additionalNotes"`
	AttachmentIDs      []string  `json:"attachmentIds"`
}

type CreateRegistrationFormResponse struct {
	ID string `json:"id"`
}

type ListRegistrationFormsRequest struct {
	Search *string `form:"search"`
}

type ListRegistrationFormsResponse struct {
	ID                    string    `json:"id"`
	FirstName             string    `json:"firstName"`
	LastName              string    `json:"lastName"`
	Bsn                   string    `json:"bsn"`
	DateOfBirth           time.Time `json:"dateOfBirth"`
	RefferingOrgID        *string   `json:"refferingOrgId"`
	OrgName               *string   `json:"orgName"`
	OrgContactPerson      *string   `json:"orgContactPerson"`
	OrgPhoneNumber        *string   `json:"orgPhoneNumber"`
	OrgEmail              *string   `json:"orgEmail"`
	CareType              string    `json:"careType"`
	RegistrationDate      time.Time `json:"registrationDate"`
	RegistrationReason    string    `json:"registrationReason"`
	AdditionalNotes       *string   `json:"additionalNotes"`
	NumberOfAttachments   int       `json:"numberOfAttachments"`
	Status                *string   `json:"status"`
}
