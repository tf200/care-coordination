package employee

import (
	"care-cordination/lib/resp"
	"context"
)

type EmployeeService interface {
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (CreateEmployeeResponse, error)
	ListEmployees(
		ctx context.Context,
		req *ListEmployeesRequest,
	) (*resp.PaginationResponse[ListEmployeesResponse], error)
	GetEmployeeByID(ctx context.Context, id string) (*GetEmployeeByIDResponse, error)
	GetMyProfile(ctx context.Context) (*GetMyProfileResponse, error)
	UpdateEmployee(ctx context.Context, id string, req *UpdateEmployeeRequest) (*UpdateEmployeeResponse, error)
	DeleteEmployee(ctx context.Context, id string) error
}
