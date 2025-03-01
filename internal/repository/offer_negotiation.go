package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type OfferNegotiation interface {
	Create(ctx context.Context, req types.OfferNegotiation) error
	FindByOfferIDAndStatus(ctx context.Context, offerID uuid.UUID, status types.OfferNegotiationStatus) (types.OfferNegotiation, error)
	FindByOfferIDsAndStatus(ctx context.Context, offerIDs []uuid.UUID, status types.OfferNegotiationStatus) ([]types.OfferNegotiation, error)
	FindAllByOfferID(ctx context.Context, offerID uuid.UUID) ([]types.OfferNegotiation, error)
}

type offerNegotiationImpl struct {
	db *sqlx.DB
}

func NewOfferNegotiation(db *sqlx.DB) OfferNegotiation {
	return &offerNegotiationImpl{
		db: db,
	}
}

func (r *offerNegotiationImpl) Create(ctx context.Context, req types.OfferNegotiation) error {
	query := `
		INSERT INTO offer_negotiations (
			id,
			offer_id,
			message,
			requested_service_cost,
			status,
			created_at
		)
		VALUES (
			:id,
			:offer_id,
			:message,
			:requested_service_cost,
			:status,
			:created_at
		)
	`

	if _, err := r.db.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *offerNegotiationImpl) FindByOfferIDAndStatus(ctx context.Context, offerID uuid.UUID, status types.OfferNegotiationStatus) (types.OfferNegotiation, error) {
	res := types.OfferNegotiation{}

	query := `
		SELECT
			id,
			offer_id,
			message,
			requested_service_cost,
			status,
			created_at
		FROM offer_negotiations
		WHERE offer_id = $1
			AND status = $2
	`

	err := r.db.GetContext(ctx, &res, query, offerID, status)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, err
	}

	return res, nil
}

func (r *offerNegotiationImpl) FindByOfferIDsAndStatus(ctx context.Context, offerIDs []uuid.UUID, status types.OfferNegotiationStatus) ([]types.OfferNegotiation, error) {
	res := []types.OfferNegotiation{}

	query := `
		SELECT
			id,
			offer_id,
			message,
			requested_service_cost,
			status,
			created_at
		FROM offer_negotiations
		WHERE offer_id = ANY($1)
			AND status = $2
	`

	if err := r.db.SelectContext(ctx, &res, query, pq.Array(offerIDs), status); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *offerNegotiationImpl) FindAllByOfferID(ctx context.Context, offerID uuid.UUID) ([]types.OfferNegotiation, error) {
	res := []types.OfferNegotiation{}

	query := `
		SELECT
			id,
			offer_id,
			message,
			requested_service_cost,
			status,
			created_at
		FROM offer_negotiations
		WHERE offer_id = $1
	`

	if err := r.db.SelectContext(ctx, &res, query, offerID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
