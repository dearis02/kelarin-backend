package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

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
	FindByServiceID(ctx context.Context, serviceID uuid.UUID) (types.ServiceProvider, error)
	FindForUpdateByID(ctx context.Context, tx dbUtil.Tx, ID uuid.UUID) (types.ServiceProvider, error)
	UpdateAsFeedbackGiven(ctx context.Context, tx dbUtil.Tx, req types.ServiceProvider) error
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

func (r *serviceProviderImpl) FindByServiceID(ctx context.Context, serviceID uuid.UUID) (types.ServiceProvider, error) {
	res := types.ServiceProvider{}

	query := `
		SELECT
			service_providers.id,
			service_providers.user_id,
			service_providers.name,
			service_providers.description,
			service_providers.has_physical_office,
			service_providers.office_coordinates,
			service_providers.address,
			service_providers.mobile_phone_number,
			service_providers.telephone,
			service_providers.logo_image,
			service_providers.received_rating_count,
			service_providers.received_rating_average,
			service_providers.credit,
			service_providers.is_deleted,
			service_providers.created_at
		FROM service_providers
		INNER JOIN services
			ON services.service_provider_id = service_providers.id
		WHERE services.id = $1
	`

	err := r.db.GetContext(ctx, &res, query, serviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
func (r serviceProviderImpl) FindForUpdateByID(ctx context.Context, _tx dbUtil.Tx, ID uuid.UUID) (types.ServiceProvider, error) {
	res := types.ServiceProvider{}

	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return res, err
	}

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
		FOR UPDATE
	`

	err = tx.GetContext(ctx, &res, query, ID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r serviceProviderImpl) UpdateAsFeedbackGiven(ctx context.Context, _tx dbUtil.Tx, req types.ServiceProvider) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		UPDATE service_providers
		SET received_rating_count = $1,
			received_rating_average = $2
		WHERE id = $3
	`

	if _, err := tx.ExecContext(ctx, query, req.ReceivedRatingCount, req.ReceivedRatingAverage, req.ID); err != nil {
		return errors.New(err)
	}

	return nil
}
