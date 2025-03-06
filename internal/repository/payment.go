package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Payment interface {
	FindByID(ctx context.Context, ID uuid.UUID) (types.Payment, error)
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.Payment) error
	UpdateStatusTx(ctx context.Context, tx *sqlx.Tx, req types.Payment) error
}

type paymentImpl struct {
	db *sqlx.DB
}

func NewPayment(db *sqlx.DB) Payment {
	return &paymentImpl{db: db}
}

func (r *paymentImpl) FindByID(ctx context.Context, ID uuid.UUID) (types.Payment, error) {
	res := types.Payment{}

	query := `
		SELECT 
			id,
			payment_method_id,
			user_id,
			amount,
			admin_fee,
			platform_fee,
			status,
			payment_link,
			created_at
		FROM payments
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &res, query, ID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *paymentImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, req types.Payment) error {
	query := `
		INSERT INTO payments (
			id,
			payment_method_id,
			user_id,
			amount,
			admin_fee,
			platform_fee,
			status,
			payment_link,
			created_at
		)
		VALUES (
			:id, 
			:payment_method_id,
			:user_id, 
			:amount,
			:admin_fee,
			:platform_fee,
			:status,
			:payment_link,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *paymentImpl) UpdateStatusTx(ctx context.Context, tx *sqlx.Tx, req types.Payment) error {
	query := `
		UPDATE payments
		SET
			status = :status
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}
