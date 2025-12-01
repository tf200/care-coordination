package registration

import "context"

type RegistrationService interface {
	CreateRegistrationForm(ctx context.Context, req *CreateRegistrationFormRequest) (*CreateRegistrationFormResponse, error)
	ListRegistrationForms(ctx context.Context) ([]ListRegistrationFormsResponse, error)
}
