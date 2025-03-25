package dbUtil

import (
	"context"
	"database/sql"

	"github.com/go-errors/errors"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

type SqlxTx func(ctx context.Context, opts *sql.TxOptions) (Tx, error)

type Tx interface {
	Commit() error
	Rollback() error
}

func NewSqlxTx(db *sqlx.DB) SqlxTx {
	return func(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
		tx, err := db.BeginTxx(ctx, opts)
		if err != nil {
			return tx, errors.Wrap(err, 1)
		}
		return tx, nil
	}
}

func CastSqlxTx(tx Tx) (*sqlx.Tx, error) {
	_tx, ok := tx.(*sqlx.Tx)
	if !ok {
		return _tx, errors.Wrap("invalid tx type", 1)
	}

	if _tx == nil {
		return _tx, errors.Wrap("tx is required", 1)
	}

	return _tx, nil
}
