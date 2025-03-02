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
	FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.OfferNegotiation, error)
	UpdateStatusTx(ctx context.Context, tx *sqlx.Tx, req types.OfferNegotiation) error
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

func (r *offerNegotiationImpl) FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.OfferNegotiation, error) {
	res := types.OfferNegotiation{}

	query := `
		SELECT
			offer_negotiations.id,
			offer_negotiations.offer_id,
			offer_negotiations.message,
			offer_negotiations.requested_service_cost,
			offer_negotiations.status,
			offer_negotiations.created_at
		FROM offer_negotiations
		INNER JOIN offers
			ON offers.id = offer_negotiations.offer_id
		WHERE offer_negotiations.id = $1
			AND offers.user_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, ID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, err
	}

	return res, nil
}

func (r *offerNegotiationImpl) UpdateStatusTx(ctx context.Context, tx *sqlx.Tx, req types.OfferNegotiation) error {
	query := `
		UPDATE offer_negotiations
		SET
			status = :status
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}
