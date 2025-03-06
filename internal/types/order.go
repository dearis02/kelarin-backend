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
	PaymentID         uuid.NullUUID   `db:"payment_id"`
	PaymentFulfilled  bool            `db:"payment_fulfilled"`
	ServiceFee        decimal.Decimal `db:"service_fee"`
	ServiceDate       time.Time       `db:"service_date"`
	ServiceTime       time.Time       `db:"service_time"`
	CreatedAt         time.Time       `db:"created_at"`
	UpdatedAt         null.Time       `db:"updated_at"`
}

type OrderWithRelations struct {
	Order
	ServiceID   uuid.UUID   `db:"service_id"`
	ServiceName string      `db:"service_name"`
	OfferStatus OfferStatus `db:"offer_status"`
	UserName    string      `db:"user_name"`
	UserEmail   string      `db:"user_email"`
}

type OrderWithUserAndServiceProvider struct {
	Order
	ServiceProviderUserID uuid.UUID `db:"service_provider_user_id"`
	ServiceProviderName   string    `db:"service_provider_name"`
	UserName              string    `db:"user_name"`
}

// endregion repo types
