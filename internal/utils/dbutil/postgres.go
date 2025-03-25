package dbUtil

import (
	"kelarin/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgres(cfg *config.PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.ConString)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleCons)
	db.SetMaxOpenConns(cfg.MaxOpenCons)

	return db, nil
}

func ClosePostgresConnection(db *sqlx.DB) error {
	return db.Close()
}
