package types

import (
	"time"

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
	CreatedAt          time.Time                       `db:"created_at"`
}

type ServiceProviderNotificationType int16

const (
	ServiceProviderNotificationTypeOfferReceived ServiceProviderNotificationType = iota + 1
	ServiceProviderNotificationTypeOfferNegotiationAccepted
	ServiceProviderNotificationTypeOfferNegotiationRejected
	ServiceProviderNotificationTypeOfferNegotiationCanceled
)

const (
	ServiceProviderNotificationTypeConsumerSettledPayment ServiceProviderNotificationType = iota + 101
)

// endregion repo types
