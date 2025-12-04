package referringOrgs

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"context"

	"go.uber.org/zap"
)

type referringOrgService struct {
	db     *db.Store
	logger *logger.Logger
}

func NewReferringOrgService(db *db.Store, logger *logger.Logger) ReferringOrgService {
	return &referringOrgService{
		db:     db,
		logger: logger,
	}
}

func (s *referringOrgService) CreateReferringOrg(ctx context.Context, req *CreateReferringOrgRequest) (*CreateReferringOrgResponse, error) {
	id := nanoid.Generate()
	err := s.db.CreateReferringOrg(ctx, db.CreateReferringOrgParams{
		ID:            id,
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		PhoneNumber:   req.PhoneNumber,
		Email:         req.Email,
	})
	if err != nil {
		s.logger.Error(ctx, "CreateReferringOrg", "Failed to create referring organization", zap.Error(err))
		return nil, ErrInternal
	}
	return &CreateReferringOrgResponse{
		ID: id,
	}, nil
}

func (s *referringOrgService) ListReferringOrgs(ctx context.Context, req *ListReferringOrgsRequest) (*resp.PaginationResponse[ListReferringOrgsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)
	referringOrgs, err := s.db.ListReferringOrgs(ctx, db.ListReferringOrgsParams{
		Limit:  limit,
		Offset: offset,
		Search: req.Search,
	})
	if err != nil {
		s.logger.Error(ctx, "ListReferringOrgs", "Failed to list referring organizations", zap.Error(err))
		return nil, ErrInternal
	}
	listReferringOrgsResponse := []ListReferringOrgsResponse{}
	for _, referringOrg := range referringOrgs {
		listReferringOrgsResponse = append(listReferringOrgsResponse, ListReferringOrgsResponse{
			ID:            referringOrg.ID,
			Name:          referringOrg.Name,
			ContactPerson: referringOrg.ContactPerson,
			PhoneNumber:   referringOrg.PhoneNumber,
			Email:         referringOrg.Email,
			CreatedAt:     referringOrg.CreatedAt.Time,
			UpdatedAt:     referringOrg.UpdatedAt.Time,
		})
	}
	totalCount := 0
	if len(referringOrgs) > 0 {
		totalCount = int(referringOrgs[0].TotalCount)
	}
	// Use page and pageSize (not offset and limit) for correct pagination metadata
	result := resp.PagRespWithParams(listReferringOrgsResponse, totalCount, page, pageSize)
	return &result, nil
}
