package types

import (
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
)

// region repo types

type ServiceProviderNotification struct {
	ID                 uuid.UUID                       `db:"id"`
	ServiceProviderID  uuid.UUID                       `db:"service_provider_id"`
	OfferID            uuid.NullUUID                   `db:"offer_id"`
	OfferNegotiationID uuid.NullUUID                   `db:"offer_negotiation_id"`
	OrderID            uuid.NullUUID                   `db:"order_id"`
	Type               ServiceProviderNotificationType `db:"type"`
	Read               bool                            `db:"read"`
	CreatedAt          time.Time                       `db:"created_at"`
}

type ServiceProviderNotificationType int16

const (
	ServiceProviderNotificationTypeOfferReceived ServiceProviderNotificationType = iota + 1
	ServiceProviderNotificationTypeOfferCanceled
	ServiceProviderNotificationTypeOfferNegotiationAccepted
	ServiceProviderNotificationTypeOfferNegotiationRejected
)

const (
	ServiceProviderNotificationTypeConsumerSettledPayment ServiceProviderNotificationType = iota + 101
)

const (
	ServiceProviderNotificationTypeOrderFinished ServiceProviderNotificationType = iota + 201
)

type ServiceProviderNotificationWithUser struct {
	ServiceProviderNotification
	UserName string `db:"user_name"`
}

// endregion repo types

// region service types

type ServiceProviderNotificationGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
	TimeZone string   `header:"Time-Zone"`
}

func (r ServiceProviderNotificationGetAllReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return nil
}

type ServiceProviderNotificationGetAllRes struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	Metadata  any       `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

type ServiceProviderNotificationMetadataOffer struct {
	OfferID uuid.UUID `json:"offer_id"`
}

type ServiceProviderNotificationMetadataOfferNegotiation struct {
	OfferNegotiationID uuid.UUID `json:"offer_negotiation_id"`
}

type ServiceProviderNotificationMetadataOrder struct {
	OrderID uuid.UUID `json:"order_id"`
}

type ServiceProviderNotificationGeneratedDetails struct {
	Title    string
	Message  string
	Metadata any
}

//endregion service types
