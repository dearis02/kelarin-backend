package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Order interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.Order) error
	FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.OrderWithRelations, error)
	UpdateAsPaymentTx(ctx context.Context, tx *sqlx.Tx, req types.Order) error
	UpdateAsPaymentFulfilledTx(ctx context.Context, tx *sqlx.Tx, req types.Order) error
	FindByPaymentID(ctx context.Context, paymentID uuid.UUID) (types.OrderWithUserAndServiceProvider, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.OrderWithServiceAndServiceProvider, error)
}

type orderImpl struct {
	db *sqlx.DB
}

func NewOrder(db *sqlx.DB) Order {
	return &orderImpl{db: db}
}

func (r *orderImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, req types.Order) error {
	query := `
		INSERT INTO orders (
			id,
			user_id,
			service_provider_id,
			offer_id,
			payment_fulfilled,
			service_fee,
			service_date,
			service_time,
			created_at
		)
		VALUES (
			:id,
			:user_id,
			:service_provider_id,
			:offer_id,
			:payment_fulfilled,
			:service_fee,
			:service_date,
			:service_time,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *orderImpl) FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.OrderWithRelations, error) {
	res := types.OrderWithRelations{}

	query := `
		SELECT
			orders.id,
			orders.user_id,
			orders.service_provider_id,
			orders.offer_id,
			orders.payment_id,
			orders.payment_fulfilled,
			orders.service_fee,
			orders.service_date,
			orders.service_time,
			orders.status,
			orders.created_at,
			orders.updated_at,
			services.id AS service_id,
			services.name AS service_name,
			offers.status AS offer_status,
			users.name AS user_name,
			users.email AS user_email
		FROM orders
		INNER JOIN users
			ON users.id = orders.user_id
		INNER JOIN offers
			ON offers.id = orders.offer_id
		INNER JOIN services
			ON services.id = offers.service_id
		WHERE orders.id = $1
			AND orders.user_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, ID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *orderImpl) UpdateAsPaymentTx(ctx context.Context, tx *sqlx.Tx, req types.Order) error {
	query := `
		UPDATE orders
		SET
			payment_id = :payment_id,
			updated_at = :updated_at
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *orderImpl) UpdateAsPaymentFulfilledTx(ctx context.Context, tx *sqlx.Tx, req types.Order) error {
	query := `
		UPDATE orders
		SET
			payment_fulfilled = :payment_fulfilled,
			updated_at = :updated_at
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *orderImpl) FindByPaymentID(ctx context.Context, paymentID uuid.UUID) (types.OrderWithUserAndServiceProvider, error) {
	res := types.OrderWithUserAndServiceProvider{}

	query := `
		SELECT
			orders.id,
			orders.user_id,
			orders.service_provider_id,
			orders.offer_id,
			orders.payment_id,
			orders.payment_fulfilled,
			orders.service_fee,
			orders.service_date,
			orders.service_time,
			orders.status,
			orders.created_at,
			orders.updated_at,
			users.name AS user_name,
			service_providers.user_id AS service_provider_user_id,
			service_providers.name AS service_provider_name
		FROM orders
		INNER JOIN users
			ON users.id = orders.user_id
		INNER JOIN service_providers
			ON service_providers.id = orders.service_provider_id
		WHERE orders.payment_id = $1
	`

	err := r.db.GetContext(ctx, &res, query, paymentID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *orderImpl) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.OrderWithServiceAndServiceProvider, error) {
	res := []types.OrderWithServiceAndServiceProvider{}

	query := `
		SELECT
			orders.id,
			orders.user_id,
			orders.service_provider_id,
			orders.offer_id,
			orders.payment_id,
			orders.payment_fulfilled,
			orders.service_fee,
			orders.service_date,
			orders.service_time,
			orders.status,
			orders.created_at,
			orders.updated_at,
			services.id AS service_id,
			services.name AS service_name,
			service_providers.name AS service_provider_name,
			service_providers.logo_image AS service_provider_logo_image
		FROM orders
		INNER JOIN offers
			ON offers.id = orders.offer_id
		INNER JOIN services
			ON services.id = offers.service_id
		INNER JOIN service_providers
			ON service_providers.id = services.service_provider_id
		WHERE orders.user_id = $1
		ORDER BY orders.id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, userID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
