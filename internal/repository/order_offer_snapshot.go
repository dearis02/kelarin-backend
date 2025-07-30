package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderOfferSnapshot interface {
	CreateTx(ctx context.Context, tx dbUtil.Tx, req types.OrderOfferSnapshot) error
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (types.OrderOfferSnapshot, error)
}

type orderOfferSnapshotImpl struct {
	db *sqlx.DB
}

func NewOrderOfferSnapshot(db *sqlx.DB) OrderOfferSnapshot {
	return &orderOfferSnapshotImpl{
		db,
	}
}

func (r *orderOfferSnapshotImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.OrderOfferSnapshot) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO order_offer_snapshots (
			order_id,
			user_address,
			service_name,
			service_delivery_methods,
			service_rules,
			service_description
		)
		VALUES (
			:order_id,
			:user_address,
			:service_name,
			:service_delivery_methods,
			:service_rules,
			:service_description
		)
	`

	_, err = tx.NamedExecContext(ctx, query, req)
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *orderOfferSnapshotImpl) FindByOrderID(ctx context.Context, orderID uuid.UUID) (types.OrderOfferSnapshot, error) {
	res := types.OrderOfferSnapshot{}

	query := `
		SELECT
			order_id,
			user_address,
			service_name,
			service_delivery_methods,
			service_rules,
			service_description
		FROM order_offer_snapshots
		WHERE order_id = $1
	`

	err := r.db.GetContext(ctx, &res, query, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
