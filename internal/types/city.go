package types

import (
	"net/http"

	"github.com/go-errors/errors"
)

// region repo types

type City struct {
	ID         int64  `db:"id"`
	ProvinceID int64  `db:"province_id"`
	Name       string `db:"name"`
}

// end of region repo types

// region service types

type CityGetByProvinceIDReq struct {
	ProvinceID int64 `json:"province_id"`
}

func (r CityGetByProvinceIDReq) Validate() error {
	if r.ProvinceID == 0 {
		return errors.New(AppErr{Code: http.StatusBadRequest, Message: "province_id is required"})
	}

	return nil
}

type CityGetAllRes struct {
	ID         int64  `json:"id"`
	ProvinceID int64  `json:"province_id"`
	Name       string `json:"name"`
}

// end of region service types
