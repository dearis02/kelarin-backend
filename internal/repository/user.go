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
	Create(ctx context.Context, user types.User) error
	CreateTx(ctx context.Context, tx *sqlx.Tx, user types.User) error
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
		return res, errors.New(err)
	}

	return res, nil
}

func (r *userImpl) Create(ctx context.Context, user types.User) error {
	statement := `
		INSERT INTO users (
			id, 
			role,
			name, 
			email, 
			password,
			auth_provider
		)
		VALUES (
			:id,
			:role,
			:name,
			:email,
			:password,
			:auth_provider
		)
	`

	_, err := r.db.NamedExecContext(ctx, statement, user)
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *userImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, user types.User) error {
	statement := `
		INSERT INTO users (
			id, 
			role,
			name, 
			email, 
			password,
			auth_provider
		)
		VALUES (
			:id,
			:role,
			:name,
			:email,
			:password,
			:auth_provider
		)
	`

	if _, err := tx.NamedExecContext(ctx, statement, user); err != nil {
		return errors.New(err)
	}

	return nil
}
