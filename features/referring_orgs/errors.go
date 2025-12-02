package referring_orgs

import "errors"

var (
	ErrReferringOrgNotFound = errors.New("referring organization not found")
	ErrInvalidRequest       = errors.New("invalid_request")
	ErrInternal             = errors.New("internal")
)
