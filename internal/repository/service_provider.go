package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ServiceProvider interface {
	Create(ctx context.Context, tx *sqlx.Tx, req types.ServiceProvider) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (types.ServiceProvider, error)
}

type serviceProviderImpl struct {
	db *sqlx.DB
}

func NewServiceProvider(db *sqlx.DB) ServiceProvider {
	return &serviceProviderImpl{db}
}

func (r *serviceProviderImpl) Create(ctx context.Context, tx *sqlx.Tx, req types.ServiceProvider) error {
	statement := `
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
	_, err := tx.NamedExecContext(ctx, statement, req)
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceProviderImpl) FindByUserID(ctx context.Context, userID uuid.UUID) (types.ServiceProvider, error) {
	res := types.ServiceProvider{}

	statement := `
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
			average_rating,
			credit,
			is_deleted,
			created_at
		FROM service_providers
		WHERE user_id = $1
	`

	err := r.db.GetContext(ctx, &res, statement, userID)
	if err != nil {
		return types.ServiceProvider{}, errors.New(err)
	}

	return res, nil
}
