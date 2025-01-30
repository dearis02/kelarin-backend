package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ChatRoom interface {
	FindByID(ctx context.Context, ID uuid.UUID) (types.ChatRoom, error)
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.ChatRoom) error
}

type chatRoomImpl struct {
	db *sqlx.DB
}

func NewChatRoom(db *sqlx.DB) ChatRoom {
	return &chatRoomImpl{db: db}
}

func (r *chatRoomImpl) FindByID(ctx context.Context, ID uuid.UUID) (types.ChatRoom, error) {
	res := types.ChatRoom{}

	statement := `
		SELECT
			id,
			service_id,
			created_at
		FROM chat_rooms
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &res, statement, ID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *chatRoomImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, req types.ChatRoom) error {
	query := `
		INSERT INTO chat_rooms (
			id,
			service_id,
			created_at
		)
		VALUES (
			:id, 
			:service_id, 
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}
