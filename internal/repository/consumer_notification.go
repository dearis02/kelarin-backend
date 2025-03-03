package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type ConsumerNotification interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.ConsumerNotification) error
}

type consumerNotificationImpl struct {
	db *sqlx.DB
}

func NewConsumerNotification(db *sqlx.DB) ConsumerNotification {
	return &consumerNotificationImpl{
		db: db,
	}
}

func (r *consumerNotificationImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, req types.ConsumerNotification) error {
	query := `
		INSERT INTO consumer_notifications (
			id,
			user_id,
			offer_negotiation_id,
			payment_id,
			order_id,
			type,
			created_at
		)
		VALUES (
			:id,
			:user_id,
			:offer_negotiation_id,
			:payment_id,
			:order_id,
			:type,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}
