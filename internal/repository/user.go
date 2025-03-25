package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type User interface {
	FindByEmail(ctx context.Context, email string) (types.User, error)
	FindByID(ctx context.Context, ID uuid.UUID) (types.User, error)
	Create(ctx context.Context, user types.User) error
	CreateTx(ctx context.Context, _tx dbUtil.Tx, user types.User) error
	FindByIDs(ctx context.Context, IDs uuid.UUIDs) ([]types.User, error)
}

type userImpl struct {
	db *sqlx.DB
}

func NewUser(db *sqlx.DB) User {
	return &userImpl{db: db}
}

func (r *userImpl) FindByID(ctx context.Context, ID uuid.UUID) (types.User, error) {
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

func (r *userImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, user types.User) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

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

func (r *userImpl) FindByIDs(ctx context.Context, IDs uuid.UUIDs) ([]types.User, error) {
	res := []types.User{}

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
		WHERE id = ANY($1)
	`

	if err := r.db.SelectContext(ctx, &res, statement, pq.Array(IDs)); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
