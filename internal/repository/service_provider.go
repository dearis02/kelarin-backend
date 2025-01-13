package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type ServiceProvider interface {
	Create(ctx context.Context, tx *sqlx.Tx, req types.ServiceProvider) error
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
			logo_image,
			created_at
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
			:logo_image,
			:created_at
		)

	`
	_, err := tx.NamedExecContext(ctx, statement, req)
	if err != nil {
		return errors.New(err)
	}

	return nil
}
