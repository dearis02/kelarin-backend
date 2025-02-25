package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Offer interface {
	Create(ctx context.Context, offer types.Offer) error
	IsPendingOfferExists(ctx context.Context, userID, serviceID uuid.UUID) (bool, error)
}

type offerImpl struct {
	db *sqlx.DB
}

func NewOffer(db *sqlx.DB) Offer {
	return &offerImpl{
		db: db,
	}
}

func (r *offerImpl) Create(ctx context.Context, offer types.Offer) error {
	query := `
		INSERT INTO offers (
			id,
			user_id,
			user_address_id,
			service_id,
			detail,
			service_cost,
			service_start_date,
			service_end_date,
			service_start_time,
			service_end_time,
			status,
			created_at
		)
		VALUES (
			:id,
			:user_id,
			:user_address_id,
			:service_id,
			:detail,
			:service_cost,
			:service_start_date,
			:service_end_date,
			:service_start_time,
			:service_end_time,
			:status,
			:created_at
		)
	`

	if _, err := r.db.NamedExecContext(ctx, query, offer); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *offerImpl) IsPendingOfferExists(ctx context.Context, userID, serviceID uuid.UUID) (bool, error) {
	query := `
		SELECT 1 
		FROM offers
		WHERE user_id = $1
			AND service_id = $2
			AND status = $3
	`

	var exs bool
	err := r.db.GetContext(ctx, &exs, query, userID, serviceID, types.OfferStatusPending)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, errors.New(err)
	}

	return exs, nil
}
