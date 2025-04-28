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

type Service interface {
	CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.Service) error
	FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Service, error)
	UpdateTx(ctx context.Context, _tx dbUtil.Tx, req types.Service) error
	FindByID(ctx context.Context, ID uuid.UUID) (types.Service, error)
	FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.Service, error)
	DeleteTx(ctx context.Context, _tx dbUtil.Tx, service types.Service) error
	FindByIDs(ctx context.Context, IDs []uuid.UUID) ([]types.Service, error)
	UpdateAsFeedbackGiven(ctx context.Context, _tx dbUtil.Tx, ID uuid.UUID, rating int16) (int32, float32, error)
}

type serviceImpl struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) Service {
	return &serviceImpl{
		db: db,
	}
}

func (r *serviceImpl) FindByID(ctx context.Context, ID uuid.UUID) (types.Service, error) {
	res := types.Service{}

	query := `
		SELECT 
			id,
			service_provider_id,
			name,
			description,
			delivery_methods,
			fee_start_at,
			fee_end_at,
			rules,
			images,
			is_available,
			received_rating_count,
			received_rating_average,
			created_at
		FROM services
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &res, query, ID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *serviceImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.Service) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	statement := `
		INSERT INTO services (
			id,
			service_provider_id,
			name,
			description,
			delivery_methods,
			fee_start_at,
			fee_end_at,
			rules,
			images,
			is_available,
			created_at
		) 
		VALUES (
			:id,
			:service_provider_id,
			:name,
			:description,
			:delivery_methods,
			:fee_start_at,
			:fee_end_at,
			:rules,
			:images,
			:is_available,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, statement, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceImpl) FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Service, error) {
	res := types.Service{}

	statement := `
		SELECT
			id,
			service_provider_id,
			name,
			description,
			delivery_methods,
			fee_start_at,
			fee_end_at,
			rules,
			images,
			is_available,
			received_rating_count,
			received_rating_average,
			created_at
		FROM services
		WHERE id = $1
		AND service_provider_id = $2
		AND is_deleted = false
	`

	err := r.db.GetContext(ctx, &res, statement, ID, serviceProviderID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *serviceImpl) UpdateTx(ctx context.Context, _tx dbUtil.Tx, req types.Service) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	statement := `
		UPDATE services
		SET
			name = :name,
			description = :description,
			delivery_methods = :delivery_methods,
			fee_start_at = :fee_start_at,
			fee_end_at = :fee_end_at,
			rules = :rules,
			images = :images,
			is_available = :is_available
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, statement, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceImpl) FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.Service, error) {
	res := []types.Service{}

	statement := `
		SELECT
			id,
			service_provider_id,
			name,
			description,
			delivery_methods,
			fee_start_at,
			fee_end_at,
			rules,
			images,
			is_available,
			received_rating_count,
			received_rating_average,
			created_at
		FROM services
		WHERE service_provider_id = $1
		AND is_deleted = false
		ORDER BY id DESC
	`

	if err := r.db.SelectContext(ctx, &res, statement, serviceProviderID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *serviceImpl) DeleteTx(ctx context.Context, _tx dbUtil.Tx, service types.Service) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	statement := `
		UPDATE services
		SET
			is_deleted = TRUE,
			deleted_at = $1
		WHERE id = $2
	`

	if _, err := tx.ExecContext(ctx, statement, service.DeletedAt, service.ID); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceImpl) FindByIDs(ctx context.Context, IDs []uuid.UUID) ([]types.Service, error) {
	res := []types.Service{}

	statement := `
		SELECT
			id,
			service_provider_id,
			name,
			description,
			delivery_methods,
			fee_start_at,
			fee_end_at,
			rules,
			images,
			is_available,
			received_rating_count,
			received_rating_average,
			created_at
		FROM services
		WHERE id = ANY($1)
		AND is_deleted = false
	`

	if err := r.db.SelectContext(ctx, &res, statement, pq.Array(IDs)); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *serviceImpl) UpdateAsFeedbackGiven(ctx context.Context, _tx dbUtil.Tx, ID uuid.UUID, rating int16) (int32, float32, error) {
	var receivedRatingCount int32
	var receivedRatingAverage float32

	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return receivedRatingCount, receivedRatingAverage, err
	}

	query := `
		SELECT id FROM services WHERE id = $1 FOR UPDATE
	`

	if _, err := tx.ExecContext(ctx, query, ID); err != nil {
		return receivedRatingCount, receivedRatingAverage, errors.New(err)
	}

	query = `
		UPDATE services
		SET
			received_rating_count = received_rating_count + 1,
			received_rating_average = ((received_rating_average * received_rating_count) + $1) / (received_rating_count + 1)
		WHERE id = $2
		RETURNING received_rating_count, received_rating_average
	`

	query = tx.Rebind(query)
	err = tx.QueryRowxContext(ctx, query, rating, ID).Scan(&receivedRatingCount, &receivedRatingAverage)
	if err != nil {
		return receivedRatingCount, receivedRatingAverage, errors.New(err)
	}

	return receivedRatingCount, receivedRatingAverage, nil
}
