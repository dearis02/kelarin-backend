package repository

import (
	"context"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ServiceProviderNotification interface {
	CreateTx(ctx context.Context, tx dbUtil.Tx, req types.ServiceProviderNotification) error
	FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.ServiceProviderNotificationWithUser, error)
}

type serviceProviderNotificationImpl struct {
	db *sqlx.DB
}

func NewServiceProviderNotification(db *sqlx.DB) ServiceProviderNotification {
	return &serviceProviderNotificationImpl{
		db,
	}
}

func (r *serviceProviderNotificationImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.ServiceProviderNotification) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

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

func (r *serviceProviderNotificationImpl) FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.ServiceProviderNotificationWithUser, error) {
	res := []types.ServiceProviderNotificationWithUser{}

	query := `
		SELECT
			service_provider_notifications.id,
			service_provider_notifications.service_provider_id,
			service_provider_notifications.offer_id,
			service_provider_notifications.offer_negotiation_id,
			service_provider_notifications.order_id,
			service_provider_notifications.type,
			service_provider_notifications.read,
			service_provider_notifications.created_at,
			users.name AS user_name
		FROM service_provider_notifications
		LEFT JOIN offer_negotiations
			ON offer_negotiations.id = service_provider_notifications.offer_negotiation_id
		LEFT JOIN offers
			ON offers.id = service_provider_notifications.offer_id
				OR offers.id = offer_negotiations.offer_id
		LEFT JOIN orders
			ON orders.id = service_provider_notifications.order_id
		LEFT JOIN users
			ON users.id = offers.user_id
				OR	users.id = orders.user_id
		WHERE service_provider_notifications.service_provider_id = $1
		ORDER BY service_provider_notifications.id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, serviceProviderID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
