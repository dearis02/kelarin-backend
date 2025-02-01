package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ServiceServiceCategory interface {
	BulkCreateTx(ctx context.Context, tx *sqlx.Tx, req []types.ServiceServiceCategory) error
	DeleteByServiceIDTx(ctx context.Context, tx *sqlx.Tx, serviceID uuid.UUID) error
}

type serviceServiceCategoryImpl struct {
	db *sqlx.DB
}

func NewServiceServiceCategory(db *sqlx.DB) ServiceServiceCategory {
	return &serviceServiceCategoryImpl{
		db: db,
	}
}

func (r *serviceServiceCategoryImpl) BulkCreateTx(ctx context.Context, tx *sqlx.Tx, req []types.ServiceServiceCategory) error {
	statement := `
		INSERT INTO service_service_categories (
			service_id,
			service_category_id
		) 
		VALUES (
			:service_id,
			:service_category_id
		)
	`

	if _, err := tx.NamedExecContext(ctx, statement, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceServiceCategoryImpl) DeleteByServiceIDTx(ctx context.Context, tx *sqlx.Tx, serviceID uuid.UUID) error {
	query := `
		DELETE FROM service_service_categories
		WHERE service_id = $1
	`

	if _, err := tx.ExecContext(ctx, query, serviceID); err != nil {
		return errors.New(err)
	}

	return nil
}
