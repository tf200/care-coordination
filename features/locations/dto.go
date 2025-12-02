package locations

type CreateLocationRequest struct {
	Name       string `json:"name" binding:"required"`
	PostalCode string `json:"postalCode" binding:"required"`
	Address    string `json:"address" binding:"required"`
	Capacity   int32  `json:"capacity" binding:"required,min=1"`
	Occupied   int32  `json:"occupied" binding:"required,min=0"`
}

type CreateLocationResponse struct {
	ID string `json:"id"`
}

type ListLocationsResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PostalCode string `json:"postalCode"`
	Address    string `json:"address"`
	Capacity   int32  `json:"capacity"`
	Occupied   int32  `json:"occupied"`
}

type ListLocationsRequest struct {
	Search *string `form:"search"`
}
