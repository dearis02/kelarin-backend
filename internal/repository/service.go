package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type Service interface {
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
