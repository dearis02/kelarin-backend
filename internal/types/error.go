package types

import "github.com/go-errors/errors"

var ErrNoData = errors.New("no data")
var ErrIDRouteParamRequired = errors.New("missing id param")
