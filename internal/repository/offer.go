package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"
	"kelarin/internal/utils"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Offer interface {
	CreateTx(ctx context.Context, tx dbUtil.Tx, offer types.Offer) error
	IsPendingOfferExists(ctx context.Context, userID, serviceID uuid.UUID) (bool, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.OfferWithServiceAndProvider, error)
	FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.Offer, error)
	FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Offer, error)
	UpdateTx(ctx context.Context, tx dbUtil.Tx, req types.Offer) error
	FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.Offer, error)
	FindForReportByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID, month, year int) (int64, []types.OfferForReport, error)
	CountGroupByStatusByServiceProviderIDAndMonthAndYear(ctx context.Context, serviceProviderID uuid.UUID, month, year int) (map[types.OfferStatus]int64, error)
	FindIDsWhereExpired(ctx context.Context, idsChan chan<- uuid.UUID) error
	UpdateAsExpired(ctx context.Context, _tx dbUtil.Tx, IDs uuid.UUIDs) error
	FindByID(ctx context.Context, ID uuid.UUID) (types.Offer, error)
}

type offerImpl struct {
	db *sqlx.DB
}

func NewOffer(db *sqlx.DB) Offer {
	return &offerImpl{
		db: db,
	}
}

func (r *offerImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, offer types.Offer) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO offers (
			id,
			user_id,
			user_address_id,
			service_id,
			detail,
			service_cost,
			service_start_date,
			service_end_date,
			service_start_time,
			service_end_time,
			status,
			created_at
		)
		VALUES (
			:id,
			:user_id,
			:user_address_id,
			:service_id,
			:detail,
			:service_cost,
			:service_start_date,
			:service_end_date,
			:service_start_time,
			:service_end_time,
			:status,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, offer); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *offerImpl) IsPendingOfferExists(ctx context.Context, userID, serviceID uuid.UUID) (bool, error) {
	query := `
		SELECT 1 
		FROM offers
		WHERE user_id = $1
			AND service_id = $2
			AND status = $3
	`

	var exs bool
	err := r.db.GetContext(ctx, &exs, query, userID, serviceID, types.OfferStatusPending)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, errors.New(err)
	}

	return exs, nil
}

func (r *offerImpl) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.OfferWithServiceAndProvider, error) {
	res := []types.OfferWithServiceAndProvider{}

	query := `
		SELECT
			offers.id,
			offers.user_id,
			offers.user_address_id,
			offers.service_id,
			offers.detail,
			offers.service_cost,
			offers.service_start_date,
			offers.service_end_date,
			offers.service_start_time,
			offers.service_end_time,
			offers.status,
			offers.created_at,
			services.name AS service_name,
			services.images[1] AS service_image,
			service_providers.id AS service_provider_id,
			service_providers.name AS service_provider_name,
			service_providers.logo_image AS service_provider_logo_image
		FROM offers
		INNER JOIN services
			ON services.id = offers.service_id
		INNER JOIN service_providers
			ON service_providers.id = services.service_provider_id
		WHERE offers.user_id = $1
		ORDER BY offers.id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, userID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *offerImpl) FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.Offer, error) {
	res := types.Offer{}

	query := `
		SELECT
			id,
			user_id,
			user_address_id,
			service_id,
			detail,
			service_cost,
			service_start_date,
			service_end_date,
			service_start_time,
			service_end_time,
			status,
			created_at
		FROM offers
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

func (r *offerImpl) FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Offer, error) {
	res := types.Offer{}

	query := `
		SELECT
			offers.id,
			offers.user_id,
			offers.user_address_id,
			offers.service_id,
			offers.detail,
			offers.service_cost,
			offers.service_start_date,
			offers.service_end_date,
			offers.service_start_time,
			offers.service_end_time,
			offers.status,
			offers.created_at
		FROM offers
		INNER JOIN services
			ON services.id = offers.service_id
		WHERE offers.id = $1
			AND services.service_provider_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, ID, serviceProviderID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *offerImpl) UpdateTx(ctx context.Context, _tx dbUtil.Tx, req types.Offer) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		UPDATE offers
		SET
			user_address_id = :user_address_id,
			detail = :detail,
			service_cost = :service_cost,
			service_start_date = :service_start_date,
			service_end_date = :service_end_date,
			service_start_time = :service_start_time,
			service_end_time = :service_end_time,
			status = :status
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *offerImpl) FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.Offer, error) {
	res := []types.Offer{}

	query := `
		SELECT
			offers.id,
			offers.user_id,
			offers.user_address_id,
			offers.service_id,
			offers.detail,
			offers.service_cost,
			offers.service_start_date,
			offers.service_end_date,
			offers.service_start_time,
			offers.service_end_time,
			offers.status,
			offers.created_at
		FROM offers
		INNER JOIN services
			ON services.id = offers.service_id
		WHERE services.service_provider_id = $1
		ORDER BY offers.id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, serviceProviderID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *offerImpl) FindForReportByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID, month, year int) (int64, []types.OfferForReport, error) {
	var total int64
	offers := []types.OfferForReport{}

	query := `
		SELECT
			DATE(offers.created_at) AS date,
			COUNT(offers.id) AS count
		FROM offers
		INNER JOIN services
			ON services.id = offers.service_id
		WHERE services.service_provider_id = $1
			AND EXTRACT(MONTH FROM offers.created_at) = $2
			AND EXTRACT(YEAR FROM offers.created_at) = $3
		GROUP BY date
		ORDER BY date
	`

	if err := r.db.SelectContext(ctx, &offers, query, serviceProviderID, month, year); err != nil {
		return total, offers, errors.New(err)
	}

	query = `
		SELECT COUNT(offers.id)
		FROM offers
		INNER JOIN services
			ON services.id = offers.service_id
		WHERE services.service_provider_id = $1
			AND EXTRACT(MONTH FROM offers.created_at) = $2
			AND EXTRACT(YEAR FROM offers.created_at) = $3
	`

	if err := r.db.GetContext(ctx, &total, query, serviceProviderID, month, year); err != nil {
		return total, offers, errors.New(err)
	}

	return total, offers, nil
}

func (r *offerImpl) CountGroupByStatusByServiceProviderIDAndMonthAndYear(ctx context.Context, serviceProviderID uuid.UUID, month, year int) (map[types.OfferStatus]int64, error) {
	res := make(map[types.OfferStatus]int64)

	query := `
		SELECT
			offers.status,
			COUNT(offers.id) AS count
		FROM offers
		INNER JOIN services
			ON services.id = offers.service_id
		WHERE services.service_provider_id = ?
			AND EXTRACT(MONTH FROM offers.created_at) = ?
			AND EXTRACT(YEAR FROM offers.created_at) = ?
		GROUP BY offers.status
	`

	query = r.db.Rebind(query)
	rows, err := r.db.QueryxContext(ctx, query, serviceProviderID, month, year)
	if err != nil {
		return res, errors.New(err)
	}

	for rows.Next() {
		var status types.OfferStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return res, errors.New(err)
		}
		res[status] = count
	}

	return res, nil
}

func (r *offerImpl) FindIDsWhereExpired(ctx context.Context, idsChan chan<- uuid.UUID) error {
	query := `
		SELECT id
		FROM offers
		WHERE service_end_date <= $1
			AND status = $2
	`

	dateNow := utils.DateNowInUTC()

	query = r.db.Rebind(query)
	rows, err := r.db.QueryxContext(ctx, query, dateNow, types.OfferStatusPending)
	if err != nil {
		return errors.New(err)
	}

	go func() {
		defer rows.Close()
		defer close(idsChan)

		for rows.Next() {
			var id uuid.UUID

			err = rows.Scan(&id)
			if err != nil {
				log.Error().Err(errors.New(err)).Send()
			}

			idsChan <- id
		}
	}()

	return nil
}

func (r *offerImpl) UpdateAsExpired(ctx context.Context, _tx dbUtil.Tx, IDs uuid.UUIDs) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `SELECT id FROM offers WHERE id = ANY($1) FOR UPDATE`

	_, err = tx.ExecContext(ctx, query, pq.Array(IDs))
	if err != nil {
		return errors.New(err)
	}

	query = `
		UPDATE offers
		SET status = $1
		WHERE id = ANY($2)
	`

	_, err = tx.ExecContext(ctx, query, types.OfferStatusExpired, pq.Array(IDs))
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *offerImpl) FindByID(ctx context.Context, ID uuid.UUID) (types.Offer, error) {
	res := types.Offer{}

	query := `
		SELECT
			id,
			user_id,
			user_address_id,
			service_id,
			detail,
			service_cost,
			service_start_date,
			service_end_date,
			service_start_time,
			service_end_time,
			status,
			created_at
		FROM offers
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
