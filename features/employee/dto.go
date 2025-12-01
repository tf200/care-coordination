package employee

import "time"

type CreateEmployeeRequest struct {
	Email       string    `json:"email" binding:"requirednm=,email"`
	Password    string    `json:"password" binding:"required"`
	FirstName   string    `json:"firstName" binding:"required"`
	LastName    string    `json:"lastName" binding:"required"`
	BSN         string    `json:"bsn" binding:"required"`
	DateOfBirth time.Time `json:"dateOfBirth" binding:"required format=2006-01-02"`
	PhoneNumber string    `json:"phoneNumber" binding:"required"`
	Gender      string    `json:"gender" binding:"required oneof=male female other"`
	Role        string    `json:"role" binding:"required"`
}

type CreateEmployeeResponse struct {
	ID string `json:"id"`
}

type ListEmployeesResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
