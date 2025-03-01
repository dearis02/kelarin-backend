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

type OfferNegotiation struct {
	ID                   uuid.UUID              `db:"id"`
	OfferID              uuid.UUID              `db:"offer_id"`
	Message              string                 `db:"message"`
	RequestedServiceCost decimal.Decimal        `db:"requested_service_cost"`
	Status               OfferNegotiationStatus `db:"status"`
	CreatedAt            time.Time              `db:"created_at"`
}

type OfferNegotiationStatus string

const (
	OfferNegotiationStatusPending  OfferNegotiationStatus = "pending"
	OfferNegotiationStatusAccepted OfferNegotiationStatus = "accepted"
	OfferNegotiationStatusRejected OfferNegotiationStatus = "rejected"
	OfferNegotiationStatusCanceled OfferNegotiationStatus = "canceled"
)

// endregion repo types

// region service types

type OfferNegotiationProviderCreateReq struct {
	AuthUser             AuthUser  `middleware:"user"`
	OfferID              uuid.UUID `json:"offer_id"`
	Message              string    `json:"message"`
	RequestedServiceCost float64   `json:"requested_service_cost"`
}

func (r OfferNegotiationProviderCreateReq) Validate(minServiceCost decimal.Decimal) error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	err := validation.ValidateStruct(&r,
		validation.Field(&r.OfferID, validation.Required),
		validation.Field(&r.RequestedServiceCost, validation.Required),
	)

	if err != nil {
		return err
	}

	ve := validation.Errors{}

	if decimal.NewFromFloat(r.RequestedServiceCost).LessThan(minServiceCost) {
		ve["requested_service_cost"] = validation.NewError("requested_service_cost_min", fmt.Sprintf("service cost must be greater than %s", minServiceCost))
	}

	if len(ve) > 0 {
		return ve
	}

	return nil
}

// endregion service types
