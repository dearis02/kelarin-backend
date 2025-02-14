package types

import (
	"time"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

const (
	ServiceProviderLogoDir = "service_provider/logo"
)

// region repo types

type ServiceProvider struct {
	ID                uuid.UUID       `db:"id"`
	UserID            uuid.UUID       `db:"user_id"`
	Name              string          `db:"name"`
	Description       string          `db:"description"`
	HasPhysicalOffice bool            `db:"has_physical_office"`
	OfficeCoordinates null.String     `db:"office_coordinates"`
	Address           string          `db:"address"`
	MobilePhoneNumber string          `db:"mobile_phone_number"`
	Telephone         string          `db:"telephone"`
	LogoImage         string          `db:"logo_image"`
	AverageRating     decimal.Decimal `db:"average_rating"`
	RatingCount       int32           `db:"rating_count"`
	Credit            decimal.Decimal `db:"credit"`
	IsDeleted         bool            `db:"is_deleted"`
	CreatedAt         time.Time       `db:"created_at"`
	DeletedAt         null.Time       `db:"deleted_at"`
}

type ServiceProviderAddress struct {
	ID      uuid.UUID `db:"id"`
	Address string    `db:"address"`
}

// end of region repo types

// region service types

type ServiceProviderCreateReq struct {
	AuthUser          AuthUser           `middleware:"user"`
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	HasPhysicalOffice bool               `json:"has_physical_office"`
	OfficeCoordinates [2]decimal.Decimal `json:"office_coordinates"`
	ProvinceID        null.Int64         `json:"province_id"`
	CityID            null.Int64         `json:"city_id"`
	Address           string             `json:"address"`
	MobilePhoneNumber string             `json:"mobile_phone_number"`
	Telephone         string             `json:"telephone"`
	Logo              string             `json:"logo"`
}

func (r ServiceProviderCreateReq) Validate() error {
	if r.AuthUser == (AuthUser{}) {
		return errors.New("AuthUser is required")
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required),
		validation.Field(&r.Description, validation.Required),
		validation.Field(&r.OfficeCoordinates, validation.Required.When(r.HasPhysicalOffice)),
		validation.Field(&r.ProvinceID, validation.Required.When(!r.HasPhysicalOffice)),
		validation.Field(&r.CityID, validation.Required.When(!r.HasPhysicalOffice)),
		validation.Field(&r.Address, validation.Required.When(!r.HasPhysicalOffice)),
		validation.Field(&r.MobilePhoneNumber, validation.Required, is.Digit),
		validation.Field(&r.Telephone, is.Digit),
		validation.Field(&r.Logo, validation.Required),
	)
}

// end of region service types
