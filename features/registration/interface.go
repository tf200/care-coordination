package registration

import (
	"care-cordination/lib/resp"
	"context"
)

type RegistrationService interface {
	CreateRegistrationForm(ctx context.Context, req *CreateRegistrationFormRequest) (*CreateRegistrationFormResponse, error)
	ListRegistrationForms(ctx context.Context, req *ListRegistrationFormsRequest) (*resp.PaginationResponse[ListRegistrationFormsResponse], error)
	UpdateRegistrationForm(ctx context.Context, id string, req *UpdateRegistrationFormRequest) (*UpdateRegistrationFormResponse, error)
	GetRegistrationForm(ctx context.Context, id string) (*GetRegistrationFormResponse, error)
	DeleteRegistrationForm(ctx context.Context, id string) (*DeleteRegistrationFormResponse, error)
}
