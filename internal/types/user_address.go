package types

import (
	"net/http"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

// region repo types

type UserAddress struct {
	ID          uuid.UUID   `db:"id"`
	Name        string      `db:"name"`
	UserID      uuid.UUID   `db:"user_id"`
	Coordinates null.String `db:"coordinates"`
	Province    string      `db:"province"`
	City        string      `db:"city"`
	Address     string      `db:"address"`
}

// endregion repo types

// region service types

type UserAddressCreateReq struct {
	AuthUser AuthUser            `middleware:"user"`
	Name     string              `json:"name"`
	Lat      decimal.NullDecimal `json:"lat"`
	Lng      decimal.NullDecimal `json:"lng"`
	Province string              `json:"province"`
	City     string              `json:"city"`
	Address  string              `json:"address"`
}

func (r UserAddressCreateReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	ve := validation.Errors{}

	if r.Lat.Valid && !r.Lng.Valid {
		ve["lng"] = validation.NewError("lng_required", "lng is required")
	}

	if r.Lng.Valid && !r.Lat.Valid {
		ve["lat"] = validation.NewError("lat_required", "lat is required")
	}

	if r.Lat.Valid && r.Lat.Decimal.LessThan(decimal.NewFromInt(-90)) || r.Lat.Decimal.GreaterThan(decimal.NewFromInt(90)) {
		ve["lat"] = validation.NewError("lat_min_max", "lat must be between -90 to 90")
	}

	if r.Lng.Valid && r.Lng.Decimal.LessThan(decimal.NewFromInt(-180)) || r.Lng.Decimal.GreaterThan(decimal.NewFromInt(180)) {
		ve["lng"] = validation.NewError("lng_min_max", "lng must be between -180 to 180")
	}

	if len(ve) > 0 {
		return ve
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&r.Province, validation.Required, validation.Length(1, 255)),
		validation.Field(&r.City, validation.Required, validation.Length(1, 255)),
		validation.Field(&r.Address, validation.Required),
	)
}

type UserAddressGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
}

func (r UserAddressGetAllReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return nil
}

type UserAddressGetAllRes struct {
	ID       uuid.UUID    `json:"id"`
	Name     string       `json:"name"`
	Lat      null.Float64 `json:"lat"`
	Lng      null.Float64 `json:"lng"`
	Province string       `json:"province"`
	City     string       `json:"city"`
	Address  string       `json:"address"`
}

type UserAddressUpdateReq struct {
	AuthUser AuthUser            `middleware:"user"`
	ID       uuid.UUID           `param:"id"`
	Name     string              `json:"name"`
	Lat      decimal.NullDecimal `json:"lat"`
	Lng      decimal.NullDecimal `json:"lng"`
	Province string              `json:"province"`
	City     string              `json:"city"`
	Address  string              `json:"address"`
}

func (r UserAddressUpdateReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.ID == uuid.Nil {
		return errors.New(AppErr{Code: http.StatusBadRequest, Message: "id param is required"})
	}

	ve := validation.Errors{}

	if r.Lat.Valid && !r.Lng.Valid {
		ve["lng"] = validation.NewError("lng_required", "lng is required")
	}

	if r.Lng.Valid && !r.Lat.Valid {
		ve["lat"] = validation.NewError("lat_required", "lat is required")
	}

	if r.Lat.Valid && r.Lat.Decimal.LessThan(decimal.NewFromInt(-90)) || r.Lat.Decimal.GreaterThan(decimal.NewFromInt(90)) {
		ve["lat"] = validation.NewError("lat_min_max", "lat must be between -90 to 90")
	}

	if r.Lng.Valid && r.Lng.Decimal.LessThan(decimal.NewFromInt(-180)) || r.Lng.Decimal.GreaterThan(decimal.NewFromInt(180)) {
		ve["lng"] = validation.NewError("lng_min_max", "lng must be between -180 to 180")
	}

	if len(ve) > 0 {
		return ve
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&r.Province, validation.Required, validation.Length(1, 255)),
		validation.Field(&r.City, validation.Required, validation.Length(1, 255)),
		validation.Field(&r.Address, validation.Required),
	)
}

// endregion service types
