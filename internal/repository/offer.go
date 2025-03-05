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
	CreateTx(ctx context.Context, tx *sqlx.Tx, offer types.Offer) error
	IsPendingOfferExists(ctx context.Context, userID, serviceID uuid.UUID) (bool, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.OfferWithServiceAndProvider, error)
	FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.Offer, error)
	FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Offer, error)
	UpdateTx(ctx context.Context, tx *sqlx.Tx, req types.Offer) error
	FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.Offer, error)
}

type offerImpl struct {
	db *sqlx.DB
}

func NewOffer(db *sqlx.DB) Offer {
	return &offerImpl{
		db: db,
	}
}

func (r *offerImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, offer types.Offer) error {
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

	if _, err := tx.NamedExecContext(ctx, query, offer); err != nil {
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

func (r *offerImpl) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.OfferWithServiceAndProvider, error) {
	res := []types.OfferWithServiceAndProvider{}

	query := `
		SELECT
			offers.id,
			offers.user_id,
			offers.user_address_id,
			offers.service_id,
			offers.detail,
			offers.service_cost,
			offers.service_start_date,
			offers.service_end_date,
			offers.service_start_time,
			offers.service_end_time,
			offers.status,
			offers.created_at,
			services.name AS service_name,
			services.images[1] AS service_image,
			service_providers.id AS service_provider_id,
			service_providers.name AS service_provider_name,
			service_providers.logo_image AS service_provider_logo_image
		FROM offers
		INNER JOIN services
			ON services.id = offers.service_id
		INNER JOIN service_providers
			ON service_providers.id = services.service_provider_id
		WHERE offers.user_id = $1
		ORDER BY offers.id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, userID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *offerImpl) FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.Offer, error) {
	res := types.Offer{}

	query := `
		SELECT
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
		FROM offers
		WHERE id = $1
			AND user_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, ID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *offerImpl) FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Offer, error) {
	res := types.Offer{}

	query := `
		SELECT
			offers.id,
			offers.user_id,
			offers.user_address_id,
			offers.service_id,
			offers.detail,
			offers.service_cost,
			offers.service_start_date,
			offers.service_end_date,
			offers.service_start_time,
			offers.service_end_time,
			offers.status,
			offers.created_at
		FROM offers
		INNER JOIN services
			ON services.id = offers.service_id
		WHERE offers.id = $1
			AND services.service_provider_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, ID, serviceProviderID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *offerImpl) UpdateTx(ctx context.Context, tx *sqlx.Tx, req types.Offer) error {
	query := `
		UPDATE offers
		SET
			user_address_id = :user_address_id,
			detail = :detail,
			service_cost = :service_cost,
			service_start_date = :service_start_date,
			service_end_date = :service_end_date,
			service_start_time = :service_start_time,
			service_end_time = :service_end_time,
			status = :status
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *offerImpl) FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.Offer, error) {
	res := []types.Offer{}

	query := `
		SELECT
			offers.id,
			offers.user_id,
			offers.user_address_id,
			offers.service_id,
			offers.detail,
			offers.service_cost,
			offers.service_start_date,
			offers.service_end_date,
			offers.service_start_time,
			offers.service_end_time,
			offers.status,
			offers.created_at
		FROM offers
		INNER JOIN services
			ON services.id = offers.service_id
		WHERE services.service_provider_id = $1
		ORDER BY offers.id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, serviceProviderID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
