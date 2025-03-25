package repository

import (
	"context"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ConsumerNotification interface {
	CreateTx(ctx context.Context, tx dbUtil.Tx, req types.ConsumerNotification) error
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.ConsumerNotificationWithServiceProviderAndPayment, error)
}

type consumerNotificationImpl struct {
	db *sqlx.DB
}

func NewConsumerNotification(db *sqlx.DB) ConsumerNotification {
	return &consumerNotificationImpl{
		db: db,
	}
}

func (r *consumerNotificationImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.ConsumerNotification) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

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

func (r *consumerNotificationImpl) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.ConsumerNotificationWithServiceProviderAndPayment, error) {
	res := []types.ConsumerNotificationWithServiceProviderAndPayment{}

	query := `
		SELECT 
			consumer_notifications.id,
			consumer_notifications.user_id,
			consumer_notifications.offer_negotiation_id,
			consumer_notifications.payment_id,
			consumer_notifications.order_id,
			consumer_notifications.type,
			consumer_notifications.read,
			consumer_notifications.created_at,
			service_providers.name AS service_provider_name,
			service_providers.logo_image AS service_provider_logo_image,
			payments.amount AS payment_amount,
			payments.admin_fee AS payment_admin_fee,
			payments.platform_fee AS payment_platform_fee,
			payment_methods.name AS payment_method_name
		FROM consumer_notifications
		LEFT JOIN offer_negotiations
			ON offer_negotiations.id = consumer_notifications.offer_negotiation_id
		LEFT JOIN offers
			ON offers.id = offer_negotiations.offer_id
		LEFT JOIN services
			ON services.id = offers.service_id
		LEFT JOIN orders
			ON orders.id = consumer_notifications.order_id
		LEFT JOIN payments
			ON payments.id = consumer_notifications.payment_id
		LEFT JOIN payment_methods
			ON payment_methods.id = payments.payment_method_id
		LEFT JOIN service_providers
			ON service_providers.id = orders.service_provider_id
				OR service_providers.id = services.service_provider_id
		WHERE consumer_notifications.user_id = $1
		ORDER BY consumer_notifications.id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, userID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
