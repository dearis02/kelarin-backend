package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type ServiceProviderArea interface {
	Create(ctx context.Context, tx *sqlx.Tx, req types.ServiceProviderArea) error
}

type serviceProviderAreaImpl struct {
	db *sqlx.DB
}

func NewServiceProviderArea(db *sqlx.DB) ServiceProviderArea {
	return &serviceProviderAreaImpl{db}
}

func (r *serviceProviderAreaImpl) Create(ctx context.Context, tx *sqlx.Tx, req types.ServiceProviderArea) error {
	statement := `
		INSERT INTO service_provider_areas(
			service_provider_id, 
			province_id,
			city_id,
			district_id
		)
		VALUES(
			:service_provider_id,
			:province_id,
			:city_id,
			:district_id
		)
	`

	if _, err := tx.NamedExecContext(ctx, statement, req); err != nil {
		return errors.New(err)
	}

	return nil
}
