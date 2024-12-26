package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type User interface {
	FindByEmail(ctx context.Context, email string) (types.User, error)
}

type userImpl struct {
	db *sqlx.DB
}

func NewUser(db *sqlx.DB) User {
	return &userImpl{db: db}
}

func (r *userImpl) FindByEmail(ctx context.Context, email string) (types.User, error) {
	res := types.User{}

	statement := ` 
		SELECT 
			id,
			name,
			email,
			password,
			role,
			is_suspended,
			suspended_count,
			suspended_from,
			suspended_to,
			is_banned,
			banned_at,
			created_at
		FROM users
		WHERE email = $1
	`

	err := r.db.GetContext(ctx, &res, statement, email)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, err
	}

	return res, nil
}
