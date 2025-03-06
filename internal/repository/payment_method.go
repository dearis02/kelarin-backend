package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PaymentMethod interface {
	FindByID(ctx context.Context, ID uuid.UUID) (types.PaymentMethod, error)
}

type paymentMethodImpl struct {
	db *sqlx.DB
}

func NewPaymentMethod(db *sqlx.DB) PaymentMethod {
	return &paymentMethodImpl{db: db}
}

func (p *paymentMethodImpl) FindByID(ctx context.Context, ID uuid.UUID) (types.PaymentMethod, error) {
	res := types.PaymentMethod{}

	query := `
		SELECT
			id,
			name,
			type,
			code,
			admin_fee,
			admin_fee_unit,
			logo,
			enabled
		FROM payment_methods
		WHERE id = $1
	`

	err := p.db.GetContext(ctx, &res, query, ID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
