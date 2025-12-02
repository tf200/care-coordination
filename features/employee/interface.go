package employee

import (
	"care-cordination/lib/resp"
	"context"
)

type EmployeeService interface {
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (CreateEmployeeResponse, error)
	ListEmployees(ctx context.Context, req *ListEmployeesRequest) (*resp.PaginationResponse[ListEmployeesResponse], error)
}
