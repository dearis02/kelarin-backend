package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type Province interface {
	FindByID(ctx context.Context, id int64) (types.Province, error)
	FindByName(ctx context.Context, name string) (types.Province, error)
	Create(ctx context.Context, req []types.Province) error
}

type provinceImpl struct {
	db *sqlx.DB
}

func NewProvince(db *sqlx.DB) Province {
	return &provinceImpl{db}
}

func (r *provinceImpl) FindByID(ctx context.Context, ID int64) (types.Province, error) {
	res := types.Province{}

	statement := `
		SELECT 
			id,
			name
		FROM provinces
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &res, statement, ID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *provinceImpl) FindByName(ctx context.Context, name string) (types.Province, error) {
	res := types.Province{}

	statement := `
		SELECT 
			id,
			name
		FROM provinces
		WHERE name ILIKE '%' || $1 || '%'
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &res, statement, name)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *provinceImpl) Create(ctx context.Context, req []types.Province) error {
	statement := `
		INSERT INTO provinces(
			id,
			name
		)
		VALUES(
			:id,
			:name
		)
	`

	if _, err := r.db.NamedExecContext(ctx, statement, req); err != nil {
		return errors.New(err)
	}

	return nil
}
