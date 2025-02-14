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

type ServiceProviderArea interface {
	Create(ctx context.Context, tx *sqlx.Tx, req types.ServiceProviderArea) error
	FindByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) (types.ServiceProviderAreaWithAreaDetail, error)
	FindByServiceProviderIDs(ctx context.Context, serviceProviderIDs []uuid.UUID) ([]types.ServiceProviderAreaWithAreaDetail, error)
}

type serviceProviderAreaImpl struct {
	db *sqlx.DB
}

func NewServiceProviderArea(db *sqlx.DB) ServiceProviderArea {
	return &serviceProviderAreaImpl{db}
}

func (r *serviceProviderAreaImpl) Create(ctx context.Context, tx *sqlx.Tx, req types.ServiceProviderArea) error {
	query := `
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

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceProviderAreaImpl) FindByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) (types.ServiceProviderAreaWithAreaDetail, error) {
	res := types.ServiceProviderAreaWithAreaDetail{}

	query := `
		SELECT
			service_provider_areas.id,
			service_provider_areas.service_provider_id,
			service_provider_areas.province_id,
			service_provider_areas.city_id,
			service_provider_areas.district_id,
			provinces.name AS province_name,
			cities.name AS city_name
		FROM service_provider_areas
		INNER JOIN provinces 
			ON service_provider_areas.province_id = provinces.id
		INNER JOIN cities 
			ON service_provider_areas.city_id = cities.id
		WHERE service_provider_areas.service_provider_id = $1
	`

	err := r.db.GetContext(ctx, &res, query, serviceProviderID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *serviceProviderAreaImpl) FindByServiceProviderIDs(ctx context.Context, IDs []uuid.UUID) ([]types.ServiceProviderAreaWithAreaDetail, error) {
	res := []types.ServiceProviderAreaWithAreaDetail{}

	query := `
		SELECT
			service_provider_areas.id,
			service_provider_areas.service_provider_id,
			service_provider_areas.province_id,
			service_provider_areas.city_id,
			service_provider_areas.district_id,
			provinces.name AS province_name,
			cities.name AS city_name
		FROM service_provider_areas
		INNER JOIN provinces 
			ON service_provider_areas.province_id = provinces.id
		INNER JOIN cities 
			ON service_provider_areas.city_id = cities.id
		WHERE service_provider_areas.service_provider_id = ANY($1)
	`

	if err := r.db.SelectContext(ctx, &res, query, pq.Array(IDs)); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
