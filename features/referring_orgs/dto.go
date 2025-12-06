package referringOrgs

import (
	"time"
)

type CreateReferringOrgRequest struct {
	Name          string `json:"name" binding:"required"`
	ContactPerson string `json:"contactPerson" binding:"required"`
	PhoneNumber   string `json:"phoneNumber" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
}

type CreateReferringOrgResponse struct {
	ID string `json:"id"`
}

type ListReferringOrgsRequest struct {
	Search *string `form:"search"`
}

type ListReferringOrgsResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ContactPerson    string    `json:"contactPerson"`
	PhoneNumber      string    `json:"phoneNumber"`
	Email            string    `json:"email"`
	InCareCount      int64     `json:"inCareCount"`
	WaitingListCount int64     `json:"waitingListCount"`
	DischargedCount  int64     `json:"dischargedCount"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
