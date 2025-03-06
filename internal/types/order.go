package types

import (
	"time"

	"github.com/go-errors/errors"
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
	Status            OrderStatus     `db:"status"`
	CreatedAt         time.Time       `db:"created_at"`
	UpdatedAt         null.Time       `db:"updated_at"`
}

type OrderStatus string

const (
	OrderStatusPending  OrderStatus = "pending"
	OrderStatusOngoing  OrderStatus = "ongoing"
	OrderStatusFinished OrderStatus = "finished"
)

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

type OrderWithServiceAndServiceProvider struct {
	Order
	ServiceID                uuid.UUID `db:"service_id"`
	ServiceName              string    `db:"service_name"`
	ServiceProviderName      string    `db:"service_provider_name"`
	ServiceProviderLogoImage string    `db:"service_provider_logo_image"`
}

// endregion repo types

// region service types

type OrderConsumerGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
	TimeZone string   `header:"Time-Zone"`
}

func (r OrderConsumerGetAllReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return nil
}

type OrderConsumerGetAllRes struct {
	ID               uuid.UUID                             `json:"id"`
	ServiceFee       decimal.Decimal                       `json:"service_fee"`
	ServiceDate      string                                `json:"service_date"`
	ServiceTime      string                                `json:"service_time"`
	PaymentFulfilled bool                                  `json:"payment_fulfilled"`
	Status           OrderStatus                           `json:"status"`
	CreatedAt        time.Time                             `json:"created_at"`
	Service          OrderConsumerGetAllResService         `json:"service"`
	ServiceProvider  OrderConsumerGetAllResServiceProvider `json:"service_provider"`
}

type OrderConsumerGetAllResService struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type OrderConsumerGetAllResServiceProvider struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	LogoURL string    `json:"logo_url"`
}

// endregion repo types
