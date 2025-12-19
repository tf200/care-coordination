package locations

import (
	"care-cordination/lib/resp"
	"context"
)

type LocationService interface {
	CreateLocation(ctx context.Context, req *CreateLocationRequest) (CreateLocationResponse, error)
	ListLocations(
		ctx context.Context,
		req *ListLocationsRequest,
	) (*resp.PaginationResponse[ListLocationsResponse], error)
}
