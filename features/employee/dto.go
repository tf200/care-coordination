package employee

import "time"

type CreateEmployeeRequest struct {
	Email         string    `json:"email"       binding:"requirednm=,email"`
	Password      string    `json:"password"    binding:"required"`
	FirstName     string    `json:"firstName"   binding:"required"`
	LastName      string    `json:"lastName"    binding:"required"`
	BSN           string    `json:"bsn"         binding:"required"`
	DateOfBirth   time.Time `json:"dateOfBirth" binding:"required,datetime=2006-01-02"`
	PhoneNumber   string    `json:"phoneNumber" binding:"required"`
	Gender        string    `json:"gender"      binding:"required oneof=male female other"`
	Role          string    `json:"role"        binding:"required"`
	LocationID    string    `json:"locationId"  binding:"required"`
	ContractHours *int32    `json:"contractHours"`
	ContractType  *string   `json:"contractType"`
}

type CreateEmployeeResponse struct {
	ID string `json:"id"`
}

type ListEmployeesResponse struct {
	ID            string  `json:"id"`
	UserID        string  `json:"userId"`
	FirstName     string  `json:"firstName"`
	LastName      string  `json:"lastName"`
	Email         string  `json:"email"`
	BSN           string  `json:"bsn"`
	DateOfBirth   string  `json:"dateOfBirth"`
	PhoneNumber   string  `json:"phoneNumber"`
	Gender        string  `json:"gender"`
	LocationID    string  `json:"locationId"`
	LocationName  string  `json:"locationName"`
	ContractHours *int32  `json:"contractHours"`
	ContractType  *string `json:"contractType"`
	ClientCount   int64   `json:"clientCount"`
}

type ListEmployeesRequest struct {
	Search *string `form:"search"`
}

type GetEmployeeByIDResponse struct {
	ID            string  `json:"id"`
	UserID        string  `json:"userId"`
	FirstName     string  `json:"firstName"`
	LastName      string  `json:"lastName"`
	Email         string  `json:"email"`
	BSN           string  `json:"bsn"`
	DateOfBirth   string  `json:"dateOfBirth"`
	PhoneNumber   string  `json:"phoneNumber"`
	Gender        string  `json:"gender"`
	LocationID    string  `json:"locationId"`
	LocationName  string  `json:"locationName"`
	ContractHours *int32  `json:"contractHours"`
	ContractType  *string `json:"contractType"`
	RoleID        *string `json:"roleId"`
	RoleName      *string `json:"roleName"`
	ClientCount   int64   `json:"clientCount"`
}

type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PermissionResponse struct {
	ID          string `json:"id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

type GetMyProfileResponse struct {
	ID          string               `json:"id"`
	UserID      string               `json:"userId"`
	FirstName   string               `json:"firstName"`
	LastName    string               `json:"lastName"`
	Email       string               `json:"email"`
	BSN         string               `json:"bsn"`
	DateOfBirth string               `json:"dateOfBirth"`
	PhoneNumber string               `json:"phoneNumber"`
	Gender      string               `json:"gender"`
	Role        string               `json:"role"`
	Permissions []PermissionResponse `json:"permissions"`
}
