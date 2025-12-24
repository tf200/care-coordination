package registration

import (
	"time"
)

type CreateRegistrationFormRequest struct {
	FirstName          string   `json:"firstName"          binding:"required"`
	LastName           string   `json:"lastName"           binding:"required"`
	BSN                string   `json:"bsn"                binding:"required"`
	DateOfBirth        string   `json:"dateOfBirth"        binding:"required"                                                                                            format:"2006-01-02"`
	PhoneNumber        *string  `json:"phoneNumber"`
	Gender             string   `json:"gender"             binding:"required,oneof=male female other"`
	RefferingOrgID     *string  `json:"refferingOrgId"     binding:"required"`
	CareType           string   `json:"careType"           binding:"required,oneof=protected_living semi_independent_living independent_assisted_living ambulatory_care"`
	RegistrationDate   string   `json:"registrationDate"   binding:"required"                                                                                            format:"2006-01-02"`
	RegistrationReason string   `json:"registrationReason" binding:"required"`
	AdditionalNotes    *string  `json:"additionalNotes"`
	AttachmentIDs      []string `json:"attachmentIds"`
}

type CreateRegistrationFormResponse struct {
	ID string `json:"id"`
}

type ListRegistrationFormsRequest struct {
	Search          *string `form:"search"`
	Status          *string `form:"status"`
	IntakeCompleted *bool   `form:"intakeCompleted"`
}

type ListRegistrationFormsResponse struct {
	ID                  string    `json:"id"`
	FirstName           string    `json:"firstName"`
	LastName            string    `json:"lastName"`
	Bsn                 string    `json:"bsn"`
	DateOfBirth         time.Time `json:"dateOfBirth"`
	RefferingOrgID      *string   `json:"refferingOrgId"`
	OrgName             *string   `json:"orgName"`
	OrgContactPerson    *string   `json:"orgContactPerson"`
	OrgPhoneNumber      *string   `json:"orgPhoneNumber"`
	OrgEmail            *string   `json:"orgEmail"`
	CareType            string    `json:"careType"`
	RegistrationDate    time.Time `json:"registrationDate"`
	RegistrationReason  string    `json:"registrationReason"`
	AdditionalNotes     *string   `json:"additionalNotes"`
	NumberOfAttachments int       `json:"numberOfAttachments"`
	Status              *string   `json:"status"`
	IntakeCompleted     bool      `json:"intakeCompleted"`
}

type UpdateRegistrationFormRequest struct {
	FirstName          *string  `json:"firstName"`
	LastName           *string  `json:"lastName"`
	BSN                *string  `json:"bsn"`
	DateOfBirth        *string  `json:"dateOfBirth"        format:"2006-01-02"`
	PhoneNumber        *string  `json:"phoneNumber"`
	Gender             *string  `json:"gender"                                 binding:"omitempty,oneof=male female other"`
	RefferingOrgID     *string  `json:"refferingOrgId"`
	CareType           *string  `json:"careType"                               binding:"omitempty,oneof=protected_living semi_independent_living independent_assisted_living ambulatory_care"`
	RegistrationDate   *string  `json:"registrationDate"   format:"2006-01-02"`
	RegistrationReason *string  `json:"registrationReason"`
	AdditionalNotes    *string  `json:"additionalNotes"`
	Status             *string  `json:"status"                                 binding:"omitempty,oneof=pending approved rejected in_review"`
	AttachmentIDs      []string `json:"attachmentIds"`
}

type UpdateRegistrationFormResponse struct {
	ID string `json:"id"`
}

type GetRegistrationFormResponse struct {
	ID                 string    `json:"id"`
	FirstName          string    `json:"firstName"`
	LastName           string    `json:"lastName"`
	Bsn                string    `json:"bsn"`
	Gender             string    `json:"gender"`
	DateOfBirth        time.Time `json:"dateOfBirth"`
	RefferingOrgID     *string   `json:"refferingOrgId"`
	OrgName            *string   `json:"orgName"`
	OrgContactPerson   *string   `json:"orgContactPerson"`
	OrgPhoneNumber     *string   `json:"orgPhoneNumber"`
	OrgEmail           *string   `json:"orgEmail"`
	CareType           string    `json:"careType"`
	RegistrationDate   time.Time `json:"registrationDate"`
	RegistrationReason string    `json:"registrationReason"`
	AdditionalNotes    *string   `json:"additionalNotes"`
	AttachmentIDs      []string  `json:"attachmentIds"`
	Status             *string   `json:"status"`
	IntakeCompleted    bool      `json:"intakeCompleted"`
	HasClient          bool      `json:"hasClient"`
}

type DeleteRegistrationFormResponse struct {
	ID string `json:"id"`
}

type GetRegistrationStatsResponse struct {
	TotalCount    int `json:"totalCount"`
	ApprovedCount int `json:"approvedCount"`
	InReviewCount int `json:"inReviewCount"`
}
