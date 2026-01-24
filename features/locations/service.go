package locations

import (
	"care-cordination/lib/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"context"

	"go.uber.org/zap"
)

type locationService struct {
	store  *db.Store
	logger logger.Logger
}

func NewLocationService(store *db.Store, logger logger.Logger) LocationService {
	return &locationService{
		store:  store,
		logger: logger,
	}
}

func (s *locationService) CreateLocation(
	ctx context.Context,
	req *CreateLocationRequest,
) (CreateLocationResponse, error) {
	id := nanoid.Generate()
	err := s.store.CreateLocation(ctx, db.CreateLocationParams{
		ID:         id,
		Name:       req.Name,
		PostalCode: req.PostalCode,
		Address:    req.Address,
		Capacity:   req.Capacity,
		Occupied:   req.Occupied,
	})
	if err != nil {
		s.logger.Error(ctx, "CreateLocation", "Failed to create location", zap.Error(err))
		return CreateLocationResponse{}, ErrInternal
	}
	return CreateLocationResponse{
		ID: id,
	}, nil
}

func (s *locationService) ListLocations(
	ctx context.Context,
	req *ListLocationsRequest,
) (*resp.PaginationResponse[ListLocationsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	locations, err := s.store.ListLocations(ctx, db.ListLocationsParams{
		Limit:  limit,
		Offset: offset,
		Search: req.Search,
	})
	if err != nil {
		s.logger.Error(ctx, "ListLocations", "Failed to list locations", zap.Error(err))
		return nil, ErrInternal
	}

	listLocationsResponse := []ListLocationsResponse{}
	totalCount := 0

	for _, location := range locations {
		listLocationsResponse = append(listLocationsResponse, ListLocationsResponse{
			ID:         location.ID,
			Name:       location.Name,
			PostalCode: location.PostalCode,
			Address:    location.Address,
			Capacity:   location.Capacity,
			Occupied:   location.Occupied,
		})
		if totalCount == 0 {
			totalCount = int(location.TotalCount)
		}
	}

	result := resp.PagRespWithParams(listLocationsResponse, totalCount, page, pageSize)
	return &result, nil
}

func (s *locationService) UpdateLocation(
	ctx context.Context,
	id string,
	req *UpdateLocationRequest,
) (UpdateLocationResponse, error) {
	err := s.store.UpdateLocation(ctx, db.UpdateLocationParams{
		ID:         id,
		Name:       req.Name,
		PostalCode: req.PostalCode,
		Address:    req.Address,
		Capacity:   req.Capacity,
		Occupied:   req.Occupied,
	})
	if err != nil {
		s.logger.Error(ctx, "UpdateLocation", "Failed to update location", zap.Error(err))
		return UpdateLocationResponse{}, ErrInternal
	}
	return UpdateLocationResponse{
		Success: true,
	}, nil
}

func (s *locationService) DeleteLocation(
	ctx context.Context,
	id string,
) (DeleteLocationResponse, error) {
	err := s.store.SoftDeleteLocation(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "DeleteLocation", "Failed to delete location", zap.Error(err))
		return DeleteLocationResponse{}, ErrInternal
	}
	return DeleteLocationResponse{
		Success: true,
	}, nil
}

func (s *locationService) GetLocationCapacityStats(
	ctx context.Context,
) (GetLocationCapacityStatsResponse, error) {
	stats, err := s.store.GetLocationCapacityStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetLocationCapacityStats", "Failed to get capacity statistics", zap.Error(err))
		return GetLocationCapacityStatsResponse{}, ErrInternal
	}

	// Type assert interface{} values to int64, then convert to int
	totalCapacity, _ := stats.TotalCapacity.(int64)
	capacityUsed, _ := stats.CapacityUsed.(int64)

	return GetLocationCapacityStatsResponse{
		TotalCapacity: int(totalCapacity),
		CapacityUsed:  int(capacityUsed),
		FreeCapacity:  int(stats.FreeCapacity),
	}, nil
}
