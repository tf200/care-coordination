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

func (s *referringOrgService) CreateReferringOrg(
	ctx context.Context,
	req *CreateReferringOrgRequest,
) (*CreateReferringOrgResponse, error) {
	id := nanoid.Generate()
	err := s.db.CreateReferringOrg(ctx, db.CreateReferringOrgParams{
		ID:            id,
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		PhoneNumber:   req.PhoneNumber,
		Email:         req.Email,
	})
	if err != nil {
		s.logger.Error(
			ctx,
			"CreateReferringOrg",
			"Failed to create referring organization",
			zap.Error(err),
		)
		return nil, ErrInternal
	}
	return &CreateReferringOrgResponse{
		ID: id,
	}, nil
}

func (s *referringOrgService) ListReferringOrgs(
	ctx context.Context,
	req *ListReferringOrgsRequest,
) (*resp.PaginationResponse[ListReferringOrgsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)
	listReferringOrgsResponse := []ListReferringOrgsResponse{}
	var totalCount int

	if req.IncludeCounts {
		// Use the query with client counts (slower but includes counts)
		referringOrgs, err := s.db.ListReferringOrgsWithCounts(
			ctx,
			db.ListReferringOrgsWithCountsParams{
				Limit:  limit,
				Offset: offset,
				Search: req.Search,
			},
		)
		if err != nil {
			s.logger.Error(
				ctx,
				"ListReferringOrgs",
				"Failed to list referring organizations with counts",
				zap.Error(err),
			)
			return nil, ErrInternal
		}
		for _, referringOrg := range referringOrgs {
			listReferringOrgsResponse = append(listReferringOrgsResponse, ListReferringOrgsResponse{
				ID:               referringOrg.ID,
				Name:             referringOrg.Name,
				ContactPerson:    referringOrg.ContactPerson,
				PhoneNumber:      referringOrg.PhoneNumber,
				Email:            referringOrg.Email,
				InCareCount:      &referringOrg.InCareCount,
				WaitingListCount: &referringOrg.WaitingListCount,
				DischargedCount:  &referringOrg.DischargedCount,
				CreatedAt:        referringOrg.CreatedAt.Time,
				UpdatedAt:        referringOrg.UpdatedAt.Time,
			})
		}
		if len(referringOrgs) > 0 {
			totalCount = int(referringOrgs[0].TotalCount)
		}
	} else {
		// Use the simple query without counts (faster)
		referringOrgs, err := s.db.ListReferringOrgs(ctx, db.ListReferringOrgsParams{
			Limit:  limit,
			Offset: offset,
			Search: req.Search,
		})
		if err != nil {
			s.logger.Error(ctx, "ListReferringOrgs", "Failed to list referring organizations", zap.Error(err))
			return nil, ErrInternal
		}
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
		if len(referringOrgs) > 0 {
			totalCount = int(referringOrgs[0].TotalCount)
		}
	}

	// Use page and pageSize (not offset and limit) for correct pagination metadata
	result := resp.PagRespWithParams(listReferringOrgsResponse, totalCount, page, pageSize)
	return &result, nil
}

func (s *referringOrgService) UpdateReferringOrg(
	ctx context.Context,
	id string,
	req *UpdateReferringOrgRequest,
) (*UpdateReferringOrgResponse, error) {
	err := s.db.UpdateReferringOrg(ctx, db.UpdateReferringOrgParams{
		ID:            id,
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		PhoneNumber:   req.PhoneNumber,
		Email:         req.Email,
	})
	if err != nil {
		s.logger.Error(
			ctx,
			"UpdateReferringOrg",
			"Failed to update referring organization",
			zap.Error(err),
		)
		return nil, ErrInternal
	}
	return &UpdateReferringOrgResponse{
		ID: id,
	}, nil
}

func (s *referringOrgService) GetReferringOrgStats(
	ctx context.Context,
) (*GetReferringOrgStatsResponse, error) {
	stats, err := s.db.GetReferringOrgStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetReferringOrgStats", "Failed to get referring org statistics", zap.Error(err))
		return nil, ErrInternal
	}

	return &GetReferringOrgStatsResponse{
		TotalOrgs:               int(stats.TotalOrgs),
		OrgsWithInCareClients:   int(stats.OrgsWithInCareClients),
		OrgsWithWaitlistClients: int(stats.OrgsWithWaitlistClients),
		TotalClientsReferred:    int(stats.TotalClientsReferred),
	}, nil
}
