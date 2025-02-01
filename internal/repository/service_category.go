package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ServiceCategory interface {
	FindByIDs(ctx context.Context, IDs []uuid.UUID) ([]types.ServiceCategory, error)
	FindByServiceIDs(ctx context.Context, IDs []uuid.UUID) ([]types.ServiceCategory, error)
}

type serviceCategoryImpl struct {
	db *sqlx.DB
}

func NewServiceCategory(db *sqlx.DB) ServiceCategory {
	return &serviceCategoryImpl{
		db: db,
	}
}

func (r *serviceCategoryImpl) FindByIDs(ctx context.Context, IDs []uuid.UUID) ([]types.ServiceCategory, error) {
	res := []types.ServiceCategory{}

	statement := `
		SELECT
			id,
			name,
			created_at
		FROM
			service_categories
		WHERE id IN(?)
	`

	query, args, err := sqlx.In(statement, IDs)
	if err != nil {
		return res, errors.New(err)
	}

	query = r.db.Rebind(query)
	if err = r.db.SelectContext(ctx, &res, query, args...); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *serviceCategoryImpl) FindByServiceIDs(ctx context.Context, serviceIDs []uuid.UUID) ([]types.ServiceCategory, error) {
	res := []types.ServiceCategory{}

	statement := `
		SELECT
			service_categories.id,
			service_categories.name,
			service_categories.created_at
		FROM
			service_categories
		INNER JOIN service_service_categories
			ON service_categories.id = service_service_categories.service_category_id
		WHERE service_service_categories.service_id IN(?)
	`

	query, args, err := sqlx.In(statement, serviceIDs)
	if err != nil {
		return res, errors.New(err)
	}

	query = r.db.Rebind(query)
	if err = r.db.SelectContext(ctx, &res, query, args...); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
