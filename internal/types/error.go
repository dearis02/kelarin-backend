package types

import (
	"net/http"

	"github.com/go-errors/errors"
)

var ErrNoData = errors.New("no data")
var ErrIDRouteParamRequired = errors.New(AppErr{Code: http.StatusBadRequest, Message: "id param is required"})

var ErrMustBeSlice = errors.New("must be slice")
var ErrMustBeStruct = errors.New("must be struct")
var ErrEmptySlice = errors.New("empty slice")
