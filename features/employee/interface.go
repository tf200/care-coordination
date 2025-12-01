package employee

import "context"

type EmployeeService interface {
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (CreateEmployeeResponse, error)
	ListEmployees(ctx context.Context) ([]ListEmployeesResponse, error)
}
