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
