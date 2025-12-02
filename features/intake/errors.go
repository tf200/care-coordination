package intake

import "errors"

var ErrInternal = errors.New("internal server error")
var ErrInvalidRequest = errors.New("invalid request")