package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

// region repo types

type Order struct {
	ID                uuid.UUID       `db:"id"`
	UserID            uuid.UUID       `db:"user_id"`
	ServiceProviderID uuid.UUID       `db:"service_provider_id"`
	OfferID           uuid.UUID       `db:"offer_id"`
	PaymentFulfilled  bool            `db:"payment_fulfilled"`
	ServiceFee        decimal.Decimal `db:"service_fee"`
	ServiceDate       time.Time       `db:"service_date"`
	ServiceTime       time.Time       `db:"service_time"`
	CreatedAt         time.Time       `db:"created_at"`
	UpdatedAt         null.Time       `db:"updated_at"`
}

// endregion repo types
