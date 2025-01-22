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
