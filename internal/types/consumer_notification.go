package types

import (
	"time"

	"github.com/google/uuid"
)

// region repo types

type ConsumerNotification struct {
	ID                 uuid.UUID                `db:"id"`
	UserID             uuid.UUID                `db:"user_id"`
	OfferNegotiationID uuid.NullUUID            `db:"offer_negotiation_id"`
	PaymentID          uuid.NullUUID            `db:"payment_id"`
	OrderID            uuid.NullUUID            `db:"order_id"`
	Type               ConsumerNotificationType `db:"type"`
	CreatedAt          time.Time                `db:"created_at"`
}

type ConsumerNotificationType int16

const (
	ConsumerNotificationTypeOfferNegotiationReceived ConsumerNotificationType = iota + 1
	ConsumerNotificationTypeOfferAccepted
	ConsumerNotificationTypeOfferRejected
)

// endregion repo types
