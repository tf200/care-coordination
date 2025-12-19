package intake

import (
	"care-cordination/lib/resp"
	"context"
)

type IntakeService interface {
	CreateIntakeForm(
		ctx context.Context,
		req *CreateIntakeFormRequest,
	) (*CreateIntakeFormResponse, error)

	ListIntakeForms(
		ctx context.Context,
		req *ListIntakeFormsRequest,
	) (*resp.PaginationResponse[ListIntakeFormsResponse], error)

	GetIntakeForm(ctx context.Context, id string) (*GetIntakeFormResponse, error)

	UpdateIntakeForm(
		ctx context.Context,
		id string,
		req *UpdateIntakeFormRequest,
	) (*UpdateIntakeFormResponse, error)
}
