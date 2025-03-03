package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
)

type ConsumerNotification interface {
	Create(ctx context.Context, tx *sqlx.Tx, req types.ConsumerNotification) error
}

type consumerNotificationImpl struct {
	userRepo                 repository.User
	consumerNotificationRepo repository.ConsumerNotification
	db                       *sqlx.DB
}

func NewConsumerNotification(userRepo repository.User, consumerNotificationRepo repository.ConsumerNotification, db *sqlx.DB) ConsumerNotification {
	return &consumerNotificationImpl{
		userRepo,
		consumerNotificationRepo,
		db,
	}
}

func (s *consumerNotificationImpl) Create(ctx context.Context, _tx *sqlx.Tx, req types.ConsumerNotification) error {
	var err error

	tx := _tx
	if _tx == nil {
		tx, err = dbUtil.NewSqlxTx(ctx, s.db, nil)
		if err != nil {
			return errors.New(err)
		}

		defer tx.Rollback()
	}

	if err = s.consumerNotificationRepo.CreateTx(ctx, tx, req); err != nil {
		return err
	}

	if _tx == nil {
		err = tx.Commit()
		if err != nil {
			return errors.New(err)
		}
	}

	return nil
}
