package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type ServiceServiceCategory interface {
	BulkCreateTx(ctx context.Context, tx *sqlx.Tx, req []types.ServiceServiceCategory) error
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
