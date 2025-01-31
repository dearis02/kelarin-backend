package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.Service) error
	FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Service, error)
	UpdateTx(ctx context.Context, tx *sqlx.Tx, req types.Service) error
}

type serviceImpl struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) Service {
	return &serviceImpl{
		db: db,
	}
}

func (r *serviceImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, req types.Service) error {
	statement := `
		INSERT INTO services (
			id,
			service_provider_id,
			name,
			description,
			delivery_methods,
			fee_start_at,
			fee_end_at,
			rules,
			is_available,
			created_at
		) 
		VALUES (
			:id,
			:service_provider_id,
			:name,
			:description,
			:delivery_methods,
			:fee_start_at,
			:fee_end_at,
			:rules,
			:is_available,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, statement, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceImpl) FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Service, error) {
	res := types.Service{}

	statement := `
		SELECT
			id,
			service_provider_id,
			name,
			description,
			delivery_methods,
			fee_start_at,
			fee_end_at,
			rules,
			is_available,
			created_at
		FROM services
		WHERE id = $1
		AND service_provider_id = $2
		AND is_deleted = false
	`

	err := r.db.GetContext(ctx, &res, statement, ID, serviceProviderID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *serviceImpl) UpdateTx(ctx context.Context, tx *sqlx.Tx, req types.Service) error {
	statement := `
		UPDATE services
		SET
			name = :name,
			description = :description,
			delivery_methods = :delivery_methods,
			fee_start_at = :fee_start_at,
			fee_end_at = :fee_end_at,
			rules = :rules,
			is_available = :is_available
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, statement, req); err != nil {
		return errors.New(err)
	}

	return nil
}
