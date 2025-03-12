package types

import (
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

// region repo types

type ConsumerNotification struct {
	ID                 uuid.UUID                `db:"id"`
	UserID             uuid.UUID                `db:"user_id"`
	OfferID            uuid.NullUUID            `db:"offer_id"`
	OfferNegotiationID uuid.NullUUID            `db:"offer_negotiation_id"`
	PaymentID          uuid.NullUUID            `db:"payment_id"`
	OrderID            uuid.NullUUID            `db:"order_id"`
	Type               ConsumerNotificationType `db:"type"`
	Read               bool                     `db:"read"`
	CreatedAt          time.Time                `db:"created_at"`
}

type ConsumerNotificationType int16

const (
	ConsumerNotificationTypeOfferNegotiationReceived ConsumerNotificationType = iota + 1
	ConsumerNotificationTypeOfferAccepted
	ConsumerNotificationTypeOfferRejected
)

const (
	ConsumerNotificationTypePaymentSuccess ConsumerNotificationType = iota + 101
	ConsumerNotificationTypePaymentExpired
)

const (
	ConsumerNotificationTypeOrderFinished ConsumerNotificationType = iota + 201
)

type ConsumerNotificationWithServiceProviderAndPayment struct {
	ConsumerNotification
	ServiceProviderName      null.String         `db:"service_provider_name"`
	ServiceProviderLogoImage null.String         `db:"service_provider_logo_image"`
	PaymentAmount            decimal.NullDecimal `db:"payment_amount"`
	PaymentAdminFee          null.Int32          `db:"payment_admin_fee"`
	PaymentPlatformFee       null.Int32          `db:"payment_platform_fee"`
	PaymentMethodName        null.String         `db:"payment_method_name"`
}

// endregion repo types

// region service types

type ConsumerNotificationGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
	TimeZone string   `header:"Time-Zone"`
}

func (r ConsumerNotificationGetAllReq) Validate() error {
	if r.AuthUser == (AuthUser{}) {
		return errors.New("AuthUser is required")
	}

	return nil
}

type ConsumerNotificationGetAllRes struct {
	ID                     uuid.UUID `json:"id"`
	Title                  string    `json:"title"`
	Message                string    `json:"message"`
	ServiceProviderLogoURL string    `json:"service_provider_logo_url"`
	Read                   bool      `json:"read"`
	Metadata               any       `json:"metadata"`
	CreatedAt              time.Time `json:"created_at"`
}

type ConsumerNotificationMetadataOffer struct {
	OfferID uuid.UUID `json:"offer_id"`
}

type ConsumerNotificationMetadataOfferNegotiation struct {
	OfferNegotiationID uuid.UUID `json:"offer_negotiation_id"`
}

type ConsumerNotificationMetadataPayment struct {
	PaymentID uuid.UUID `json:"payment_id"`
}

type ConsumerNotificationMetadataOrder struct {
	OrderID uuid.UUID `json:"order_id"`
}

type ConsumerNotificationGeneratedDetails struct {
	Title    string
	Message  string
	Metadata any
}

// endregion service types
