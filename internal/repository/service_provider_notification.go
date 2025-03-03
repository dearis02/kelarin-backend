package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type ServiceProviderNotification interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.ServiceProviderNotification) error
}

type serviceProviderNotificationImpl struct {
	db *sqlx.DB
}

func NewServiceProviderNotification(db *sqlx.DB) ServiceProviderNotification {
	return &serviceProviderNotificationImpl{
		db,
	}
}

func (r *serviceProviderNotificationImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, req types.ServiceProviderNotification) error {
	query := `
		INSERT INTO service_provider_notifications (
			id,
			service_provider_id,
			offer_id,
			offer_negotiation_id,
			order_id,
			type,
			created_at
		)
		VALUES (
			:id,
			:service_provider_id,
			:offer_id,
			:offer_negotiation_id,
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
