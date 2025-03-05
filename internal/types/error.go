package types

import (
	"net/http"

	"github.com/go-errors/errors"
)

var ErrNoData = errors.New("no data")
var ErrIDRouteParamRequired = errors.New(AppErr{Code: http.StatusBadRequest, Message: "id param is required"})
