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
	FindByID(ctx context.Context, ID uuid.UUID) (types.Service, error)
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.Service) error
}

type serviceImpl struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) Service {
	return &serviceImpl{
		db: db,
	}
}

func (r *serviceImpl) FindByID(ctx context.Context, ID uuid.UUID) (types.Service, error) {
	res := types.Service{}

	query := `
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
	`

	err := r.db.GetContext(ctx, &res, query, ID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
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
