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

type ServiceFeedback interface {
	CreateTx(ctx context.Context, tx dbUtil.Tx, req types.ServiceFeedback) error
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (types.ServiceFeedback, error)
	FindByServiceIDWithUser(ctx context.Context, serviceID uuid.UUID) ([]types.ServiceFeedbackWithUser, error)
}

type serviceFeedbackImpl struct {
	db *sqlx.DB
}

func NewServiceFeedback(db *sqlx.DB) ServiceFeedback {
	return &serviceFeedbackImpl{
		db: db,
	}
}

func (s *serviceFeedbackImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.ServiceFeedback) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO service_feedbacks (
			id,
			service_id,
			order_id,
			rating,
			comment,
			created_at
		)
		VALUES (
			:id,
			:service_id,
			:order_id,
			:rating,
			:comment,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (s *serviceFeedbackImpl) FindByOrderID(ctx context.Context, orderID uuid.UUID) (types.ServiceFeedback, error) {
	res := types.ServiceFeedback{}

	query := `
		SELECT
			id,
			service_id,
			order_id,
			rating,
			comment,
			created_at
		FROM service_feedbacks
		WHERE order_id = $1
	`

	err := s.db.GetContext(ctx, &res, query, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (s *serviceFeedbackImpl) FindByServiceIDWithUser(ctx context.Context, serviceID uuid.UUID) ([]types.ServiceFeedbackWithUser, error) {
	res := []types.ServiceFeedbackWithUser{}

	query := `
		SELECT
			service_feedbacks.id,
			service_feedbacks.service_id,
			service_feedbacks.order_id,
			service_feedbacks.rating,
			service_feedbacks.comment,
			service_feedbacks.created_at,
			users.name AS user_name
		FROM service_feedbacks
		INNER JOIN orders
			ON  orders.id = service_feedbacks.order_id
		INNER JOIN users
			ON users.id = orders.user_id
		WHERE
			service_feedbacks.service_id = $1
	`

	err := s.db.SelectContext(ctx, &res, query, serviceID)
	if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
