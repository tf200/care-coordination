package referringOrgs

import (
	"care-cordination/lib/resp"
	"context"
)

type ReferringOrgService interface {
	CreateReferringOrg(
		ctx context.Context,
		req *CreateReferringOrgRequest,
	) (*CreateReferringOrgResponse, error)
	ListReferringOrgs(
		ctx context.Context,
		req *ListReferringOrgsRequest,
	) (*resp.PaginationResponse[ListReferringOrgsResponse], error)
	UpdateReferringOrg(
		ctx context.Context,
		id string,
		req *UpdateReferringOrgRequest,
	) (*UpdateReferringOrgResponse, error)
}
