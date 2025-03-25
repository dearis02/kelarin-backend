package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type Order interface {
	CreateTx(ctx context.Context, tx dbUtil.Tx, req types.Order) error
	FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.OrderWithRelations, error)
	UpdateAsPaymentTx(ctx context.Context, tx dbUtil.Tx, req types.Order) error
	UpdateAsPaymentFulfilledTx(ctx context.Context, tx dbUtil.Tx, req types.Order) error
	FindByPaymentID(ctx context.Context, paymentID uuid.UUID) (types.OrderWithUserAndServiceProvider, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.OrderWithServiceAndServiceProvider, error)
	FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.Order, error)
	FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Order, error)
	UpdateStatusTx(ctx context.Context, tx dbUtil.Tx, req types.Order) error
	FindForReportByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID, month, year int) (int64, []types.OrderForReport, error)
	FindTotalServiceFeeByServiceProviderIDAndStatusAndMonthAndYear(ctx context.Context, serviceProviderID uuid.UUID, status types.OrderStatus, month, year int) (decimal.Decimal, error)
	FindForReportExportByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.OrderForReportExport, error)
	FindByOfferID(ctx context.Context, offerID uuid.UUID) (types.OrderWithServiceAndServiceProvider, error)
}

type orderImpl struct {
	db *sqlx.DB
}

func NewOrder(db *sqlx.DB) Order {
	return &orderImpl{db: db}
}

func (r *orderImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.Order) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO orders (
			id,
			user_id,
			service_provider_id,
			offer_id,
			payment_fulfilled,
			service_fee,
			service_date,
			service_time,
			created_at
		)
		VALUES (
			:id,
			:user_id,
			:service_provider_id,
			:offer_id,
			:payment_fulfilled,
			:service_fee,
			:service_date,
			:service_time,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *orderImpl) FindByIDAndUserID(ctx context.Context, ID, userID uuid.UUID) (types.OrderWithRelations, error) {
	res := types.OrderWithRelations{}

	query := `
		SELECT
			orders.id,
			orders.user_id,
			orders.service_provider_id,
			orders.offer_id,
			orders.payment_id,
			orders.payment_fulfilled,
			orders.service_fee,
			orders.service_date,
			orders.service_time,
			orders.status,
			orders.created_at,
			orders.updated_at,
			services.id AS service_id,
			services.name AS service_name,
			offers.status AS offer_status,
			users.name AS user_name,
			users.email AS user_email
		FROM orders
		INNER JOIN users
			ON users.id = orders.user_id
		INNER JOIN offers
			ON offers.id = orders.offer_id
		INNER JOIN services
			ON services.id = offers.service_id
		WHERE orders.id = $1
			AND orders.user_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, ID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *orderImpl) UpdateAsPaymentTx(ctx context.Context, _tx dbUtil.Tx, req types.Order) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		UPDATE orders
		SET
			payment_id = :payment_id,
			updated_at = :updated_at
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *orderImpl) UpdateAsPaymentFulfilledTx(ctx context.Context, _tx dbUtil.Tx, req types.Order) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		UPDATE orders
		SET
			payment_fulfilled = :payment_fulfilled,
			updated_at = :updated_at
		WHERE id = :id
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *orderImpl) FindByPaymentID(ctx context.Context, paymentID uuid.UUID) (types.OrderWithUserAndServiceProvider, error) {
	res := types.OrderWithUserAndServiceProvider{}

	query := `
		SELECT
			orders.id,
			orders.user_id,
			orders.service_provider_id,
			orders.offer_id,
			orders.payment_id,
			orders.payment_fulfilled,
			orders.service_fee,
			orders.service_date,
			orders.service_time,
			orders.status,
			orders.created_at,
			orders.updated_at,
			users.name AS user_name,
			service_providers.user_id AS service_provider_user_id,
			service_providers.name AS service_provider_name
		FROM orders
		INNER JOIN users
			ON users.id = orders.user_id
		INNER JOIN service_providers
			ON service_providers.id = orders.service_provider_id
		WHERE orders.payment_id = $1
	`

	err := r.db.GetContext(ctx, &res, query, paymentID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *orderImpl) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]types.OrderWithServiceAndServiceProvider, error) {
	res := []types.OrderWithServiceAndServiceProvider{}

	query := `
		SELECT
			orders.id,
			orders.user_id,
			orders.service_provider_id,
			orders.offer_id,
			orders.payment_id,
			orders.payment_fulfilled,
			orders.service_fee,
			orders.service_date,
			orders.service_time,
			orders.status,
			orders.created_at,
			orders.updated_at,
			services.id AS service_id,
			services.name AS service_name,
			service_providers.name AS service_provider_name,
			service_providers.logo_image AS service_provider_logo_image,
			payment_methods.name AS payment_method_name,
			payments.amount AS payment_amount,
			payments.admin_fee AS payment_admin_fee,
			payments.platform_fee AS payment_platform_fee,
			payments.payment_link AS payment_payment_link,
			payments.status  AS payment_status
		FROM orders
		INNER JOIN offers
			ON offers.id = orders.offer_id
		INNER JOIN services
			ON services.id = offers.service_id
		INNER JOIN service_providers
			ON service_providers.id = services.service_provider_id
		LEFT JOIN payments
			ON payments.id = orders.payment_id
		LEFT JOIN payment_methods
			ON payment_methods.id = payments.payment_method_id
		WHERE orders.user_id = $1
		ORDER BY orders.id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, userID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *orderImpl) FindAllByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.Order, error) {
	res := []types.Order{}

	query := `
		SELECT
			id,
			user_id,
			service_provider_id,
			offer_id,
			payment_id,
			payment_fulfilled,
			service_fee,
			service_date,
			service_time,
			status,
			created_at,
			updated_at
		FROM orders
		WHERE service_provider_id = $1
		ORDER BY id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, serviceProviderID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *orderImpl) FindByIDAndServiceProviderID(ctx context.Context, ID, serviceProviderID uuid.UUID) (types.Order, error) {
	res := types.Order{}

	query := `
		SELECT
			id,
			user_id,
			service_provider_id,
			offer_id,
			payment_id,
			payment_fulfilled,
			service_fee,
			service_date,
			service_time,
			status,
			created_at,
			updated_at
		FROM orders
		WHERE id = $1 
			AND service_provider_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, ID, serviceProviderID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *orderImpl) UpdateStatusTx(ctx context.Context, _tx dbUtil.Tx, req types.Order) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		UPDATE orders
		SET
			status = :status,
			updated_at = :updated_at
		WHERE id = :id
	`

	_, err = tx.NamedExecContext(ctx, query, req)
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *orderImpl) FindForReportByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID, month, year int) (int64, []types.OrderForReport, error) {
	var count int64
	orders := []types.OrderForReport{}

	query := `
		SELECT
			DATE(created_at) AS date,
			COUNT(id) AS count
		FROM orders
		WHERE service_provider_id = $1
			AND EXTRACT(MONTH FROM created_at) = $2
			AND EXTRACT(YEAR FROM created_at) = $3
		GROUP BY date
		ORDER BY date
	`

	if err := r.db.SelectContext(ctx, &orders, query, serviceProviderID, month, year); err != nil {
		return count, orders, errors.New(err)
	}

	query = `
		SELECT
			COUNT(id)
		FROM orders
		WHERE service_provider_id = $1
			AND EXTRACT(MONTH FROM created_at) = $2
			AND EXTRACT(YEAR FROM created_at) = $3
	`

	if err := r.db.GetContext(ctx, &count, query, serviceProviderID, month, year); err != nil {
		return count, orders, errors.New(err)
	}

	return count, orders, nil
}

func (r *orderImpl) FindTotalServiceFeeByServiceProviderIDAndStatusAndMonthAndYear(ctx context.Context, serviceProviderID uuid.UUID, status types.OrderStatus, month, year int) (decimal.Decimal, error) {
	var total decimal.NullDecimal

	query := `
		SELECT
			SUM(service_fee)
		FROM orders
		WHERE service_provider_id = $1
			AND status = $2
			AND EXTRACT(MONTH FROM created_at) = $3
			AND EXTRACT(YEAR FROM created_at) = $4
	`

	if err := r.db.GetContext(ctx, &total, query, serviceProviderID, status, month, year); err != nil {
		return total.Decimal, errors.New(err)
	}

	return total.Decimal, nil
}

func (r *orderImpl) FindForReportExportByServiceProviderID(ctx context.Context, serviceProviderID uuid.UUID) ([]types.OrderForReportExport, error) {
	res := []types.OrderForReportExport{}

	query := `
		SELECT
			orders.id,
			orders.service_fee,
			orders.service_date,
			orders.service_time,
			orders.status,
			orders.payment_fulfilled,
			users.name AS user_name,
			users.email AS user_email,
			user_addresses.province AS user_province,
			user_addresses.city AS user_city,
			user_addresses.address AS user_address,
			orders.created_at
		FROM orders
		INNER JOIN offers
			ON offers.id = orders.offer_id
		INNER JOIN user_addresses 
			ON offers.user_address_id = user_addresses.id
		INNER JOIN users 
			ON orders.user_id = users.id
		WHERE orders.service_provider_id = $1
		GROUP BY orders.id, users.name, users.email, user_addresses.province, user_addresses.city, user_addresses.address
		ORDER BY orders.id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, serviceProviderID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *orderImpl) FindByOfferID(ctx context.Context, offerID uuid.UUID) (types.OrderWithServiceAndServiceProvider, error) {
	res := types.OrderWithServiceAndServiceProvider{}

	query := `
		SELECT
			orders.id,
			orders.user_id,
			orders.service_provider_id,
			orders.offer_id,
			orders.payment_id,
			orders.payment_fulfilled,
			orders.service_fee,
			orders.service_date,
			orders.service_time,
			orders.status,
			orders.created_at,
			orders.updated_at,
			services.id AS service_id,
			services.name AS service_name,
			service_providers.name AS service_provider_name,
			service_providers.logo_image AS service_provider_logo_image,
			payment_methods.name AS payment_method_name,
			payments.amount AS payment_amount,
			payments.admin_fee AS payment_admin_fee,
			payments.platform_fee AS payment_platform_fee,
			payments.payment_link AS payment_payment_link,
			payments.status  AS payment_status
		FROM orders
		INNER JOIN offers
			ON offers.id = orders.offer_id
		INNER JOIN services
			ON services.id = offers.service_id
		INNER JOIN service_providers
			ON service_providers.id = services.service_provider_id
		LEFT JOIN payments
			ON payments.id = orders.payment_id
		LEFT JOIN payment_methods
			ON payment_methods.id = payments.payment_method_id
		WHERE orders.offer_id = $1
		ORDER BY orders.id DESC
	`

	err := r.db.GetContext(ctx, &res, query, offerID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
