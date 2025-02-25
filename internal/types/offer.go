package types

import (
	"fmt"
	"time"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// region repo types

type Offer struct {
	ID               uuid.UUID       `db:"id"`
	UserID           uuid.UUID       `db:"user_id"`
	UserAddressID    uuid.UUID       `db:"user_address_id"`
	ServiceID        uuid.UUID       `db:"service_id"`
	Detail           string          `db:"detail"`
	ServiceCost      decimal.Decimal `db:"service_cost"`
	ServiceStartDate time.Time       `db:"service_start_date"`
	ServiceEndDate   time.Time       `db:"service_end_date"`
	ServiceStartTime string          `db:"service_start_time"`
	ServiceEndTime   string          `db:"service_end_time"`
	Status           OfferStatus     `db:"status"`
	CreatedAt        time.Time       `db:"created_at"`
}

type OfferStatus string

const (
	OfferStatusPending  OfferStatus = "pending"
	OfferStatusAccepted OfferStatus = "accepted"
	OfferStatusRejected OfferStatus = "rejected"
	OfferStatusCanceled OfferStatus = "canceled"
)

// endregion repo types

// region service types

type OfferConsumerCreateReq struct {
	AuthUser         AuthUser  `middleware:"user"`
	ServiceID        uuid.UUID `json:"service_id"`
	AddressID        uuid.UUID `json:"address_id"`
	Detail           string    `json:"detail"`
	ServiceCost      float64   `json:"service_cost"`
	ServiceStartDate string    `json:"service_start_date"`
	ServiceEndDate   string    `json:"service_end_date"`
	ServiceStartTime string    `json:"service_start_time"`
	ServiceEndTime   string    `json:"service_end_time"`
}

func (r OfferConsumerCreateReq) Validate(serviceFeeStart decimal.Decimal) error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	err := validation.ValidateStruct(&r,
		validation.Field(&r.ServiceID, validation.Required),
		validation.Field(&r.AddressID, validation.Required),
		validation.Field(&r.Detail, validation.Required),
		validation.Field(&r.ServiceCost, validation.Required),
		validation.Field(&r.ServiceStartDate, validation.Required, validation.Date(time.DateOnly)),
		validation.Field(&r.ServiceEndDate, validation.Required, validation.Date(time.DateOnly)),
		validation.Field(&r.ServiceStartTime, validation.Required, validation.Date(time.TimeOnly)),
		validation.Field(&r.ServiceEndTime, validation.Required, validation.Date(time.TimeOnly)),
	)

	if err != nil {
		return err
	}

	ve := validation.Errors{}

	now := time.Now()

	startDate, err := time.Parse(time.DateOnly, r.ServiceStartDate)
	if err != nil {
		return errors.New(err)
	}

	endDate, err := time.Parse(time.DateOnly, r.ServiceEndDate)
	if err != nil {
		return errors.New(err)
	}

	if endDate.Before(startDate) {
		ve["service_end_date"] = validation.NewError("service_end_date_min", "service_end_date must be equal or greater than service_start_date")
	}

	startTime, err := time.Parse(time.TimeOnly, r.ServiceStartTime)
	if err != nil {
		return errors.New(err)
	}

	endTime, err := time.Parse(time.TimeOnly, r.ServiceEndTime)
	if err != nil {
		return errors.New(err)
	}

	if startTime.After(endTime) || startTime.Equal(endTime) {
		ve["service_end_time"] = validation.NewError("service_end_time_min", "service_end_time must be greater than service_start_time")
		return ve
	}

	if startDate.Equal(now.Truncate(24 * time.Hour)) {
		if startTime.Before(now.Truncate(1 * time.Hour)) {
			ve["service_start_time"] = validation.NewError("service_start_time_min", "service_start_time min less than 1 hour from now")
		}
	}

	if !serviceFeeStart.LessThan(decimal.NewFromFloat(r.ServiceCost)) {
		ve["service_cost"] = validation.NewError("service_cost_min", fmt.Sprintf("service_cost must be greater or equal than %s", serviceFeeStart))
	}

	if len(ve) > 0 {
		return ve
	}

	return nil
}

// endregion service types
