package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type Order interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.Order) error
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
