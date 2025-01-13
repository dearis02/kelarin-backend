package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type City interface {
	FindByIDandProvinceID(ctx context.Context, ID, provinceID int64) (types.City, error)
	FindByProvinceIDAndName(ctx context.Context, provinceID int64, name string) (types.City, error)
}

type cityImpl struct {
	db *sqlx.DB
}

func NewCity(db *sqlx.DB) City {
	return &cityImpl{db}
}

func (r *cityImpl) FindByIDandProvinceID(ctx context.Context, ID, provinceID int64) (types.City, error) {
	res := types.City{}

	query := `
		SELECT
			id,
			province_id,
			name
		FROM cities
		WHERE id = $1
			AND province_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, ID, provinceID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, err
}

func (r *cityImpl) FindByProvinceIDAndName(ctx context.Context, provinceID int64, name string) (types.City, error) {
	res := types.City{}

	query := `
		SELECT 
			id,
			province_id,
			name
		FROM cities
		WHERE province_id = $1 
			AND name ILIKE '%' || $2 || '%'
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &res, query, provinceID, name)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, err
}
