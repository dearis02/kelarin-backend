package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserAddress interface {
	FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.UserAddress, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]types.UserAddress, error)
	Create(ctx context.Context, address types.UserAddress) error
	Update(ctx context.Context, address types.UserAddress) error
}

type userAddressImpl struct {
	db *sqlx.DB
}

func NewUserAddress(db *sqlx.DB) UserAddress {
	return &userAddressImpl{
		db: db,
	}
}

func (r *userAddressImpl) FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.UserAddress, error) {
	res := types.UserAddress{}

	query := `
		SELECT
			id,
			user_id,
			name,
			coordinates,
			province,
			city,
			detail
		FROM user_addresses
		WHERE id = $1
			AND user_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, ID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *userAddressImpl) Create(ctx context.Context, address types.UserAddress) error {
	query := `
		INSERT INTO user_addresses(
			id,
			user_id,
			name,
			coordinates,
			province,
			city,
			detail
		)
		VALUES(
			:id,
			:user_id,
			:name,
			:coordinates,
			:province,
			:city,
			:detail
		)
	`

	if _, err := r.db.NamedExecContext(ctx, query, address); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *userAddressImpl) FindByUserID(ctx context.Context, userID uuid.UUID) ([]types.UserAddress, error) {
	res := []types.UserAddress{}

	query := `
		SELECT
			id,
			user_id,
			name,
			coordinates,
			province,
			city,
			detail
		FROM user_addresses
		WHERE user_id = $1
		ORDER BY id DESC
	`

	err := r.db.SelectContext(ctx, &res, query, userID)
	if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *userAddressImpl) Update(ctx context.Context, address types.UserAddress) error {
	query := `
		UPDATE user_addresses
		SET
			name = :name,
			coordinates = :coordinates,
			province = :province,
			city = :city,
			detail = :detail
		WHERE id = :id
	`

	if _, err := r.db.NamedExecContext(ctx, query, address); err != nil {
		return errors.New(err)
	}

	return nil
}
