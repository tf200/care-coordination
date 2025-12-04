package locTransfer

type RegisterLocationTransferRequest struct {
	ClientID         string `json:"clientId" binding:"required"`
	NewLocationID    string `json:"newLocationId" binding:"required"`
	NewCoordinatorID string `json:"newCoordinatorId" binding:"required"`
}

type RegisterLocationTransferResponse struct {
	TransferID string `json:"transferId"`
}

type ListLocationTransfersRequest struct {
	Search *string `form:"search"`
}

type ListLocationTransfersResponse struct {
	ID                          string  `json:"id"`
	ClientID                    string  `json:"clientId"`
	FromLocationID              *string `json:"fromLocationId"`
	ToLocationID                string  `json:"toLocationId"`
	CurrentCoordinatorID        string  `json:"currentCoordinatorId"`
	NewCoordinatorID            string  `json:"newCoordinatorId"`
	TransferDate                string  `json:"transferDate"` // or time.Time, but for JSON, string
	Reason                      *string `json:"reason"`
	ClientFirstName             string  `json:"clientFirstName"`
	ClientLastName              string  `json:"clientLastName"`
	FromLocationName            *string `json:"fromLocationName"`
	ToLocationName              *string `json:"toLocationName"`
	CurrentCoordinatorFirstName *string `json:"currentCoordinatorFirstName"`
	CurrentCoordinatorLastName  *string `json:"currentCoordinatorLastName"`
	NewCoordinatorFirstName     *string `json:"newCoordinatorFirstName"`
	NewCoordinatorLastName      *string `json:"newCoordinatorLastName"`
}
