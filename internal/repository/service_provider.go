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

type ServiceProvider interface {
	Create(ctx context.Context, tx *sqlx.Tx, req types.ServiceProvider) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (types.ServiceProvider, error)
	FindByIDs(ctx context.Context, IDs []uuid.UUID) ([]types.ServiceProvider, error)
	FindByID(ctx context.Context, ID uuid.UUID) (types.ServiceProvider, error)
	UpdateCreditTx(ctx context.Context, req types.ServiceProvider) error
	FindByUserIDs(ctx context.Context, IDs []uuid.UUID) ([]types.ServiceProvider, error)
}

type serviceProviderImpl struct {
	db *sqlx.DB
}

func NewServiceProvider(db *sqlx.DB) ServiceProvider {
	return &serviceProviderImpl{db}
}

func (r *serviceProviderImpl) Create(ctx context.Context, tx *sqlx.Tx, req types.ServiceProvider) error {
	query := `
		INSERT INTO service_providers (
			id,
			user_id,
			name,
			description,
			has_physical_office,
			office_coordinates,
			address,
			mobile_phone_number,
			telephone,
			logo_image
		)
		VALUES (
			:id,
			:user_id,
			:name,
			:description,
			:has_physical_office,
			:office_coordinates,
			:address,
			:mobile_phone_number,
			:telephone,
			:logo_image
		)

	`
	_, err := tx.NamedExecContext(ctx, query, req)
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceProviderImpl) FindByUserID(ctx context.Context, userID uuid.UUID) (types.ServiceProvider, error) {
	res := types.ServiceProvider{}

	query := `
		SELECT
			id,
			user_id,
			name,
			description,
			has_physical_office,
			office_coordinates,
			address,
			mobile_phone_number,
			telephone,
			logo_image,
			received_rating_count,
			received_rating_average,
			credit,
			is_deleted,
			created_at
		FROM service_providers
		WHERE user_id = $1
	`

	err := r.db.GetContext(ctx, &res, query, userID)
	if err != nil {
		return types.ServiceProvider{}, errors.New(err)
	}

	return res, nil
}

func (r *serviceProviderImpl) FindByIDs(ctx context.Context, IDs []uuid.UUID) ([]types.ServiceProvider, error) {
	res := []types.ServiceProvider{}

	query := `
		SELECT
			id,
			user_id,
			name,
			description,
			has_physical_office,
			office_coordinates,
			address,
			mobile_phone_number,
			telephone,
			logo_image,
			received_rating_count,
			received_rating_average,
			credit,
			is_deleted,
			created_at
		FROM service_providers
		WHERE id = ANY($1)
		ORDER BY id DESC
	`

	err := r.db.SelectContext(ctx, &res, query, pq.Array(IDs))
	if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *serviceProviderImpl) FindByID(ctx context.Context, ID uuid.UUID) (types.ServiceProvider, error) {
	res := types.ServiceProvider{}

	query := `
		SELECT
			id,
			user_id,
			name,
			description,
			has_physical_office,
			office_coordinates,
			address,
			mobile_phone_number,
			telephone,
			logo_image,
			received_rating_count,
			received_rating_average,
			credit,
			is_deleted,
			created_at
		FROM service_providers
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

func (r *serviceProviderImpl) UpdateCreditTx(ctx context.Context, req types.ServiceProvider) error {
	query := `
		UPDATE service_providers
		SET credit = $1
		WHERE id = $2
	`

	if _, err := r.db.ExecContext(ctx, query, req.Credit, req.ID); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceProviderImpl) FindByUserIDs(ctx context.Context, IDs []uuid.UUID) ([]types.ServiceProvider, error) {
	res := []types.ServiceProvider{}

	query := `
		SELECT
			id,
			user_id,
			name,
			description,
			has_physical_office,
			office_coordinates,
			address,
			mobile_phone_number,
			telephone,
			logo_image,
			received_rating_count,
			received_rating_average,
			credit,
			is_deleted,
			created_at
		FROM service_providers
		WHERE user_id = ANY($1)
		ORDER BY id DESC
	`

	err := r.db.SelectContext(ctx, &res, query, pq.Array(IDs))
	if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
