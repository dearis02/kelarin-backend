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

type ChatRoom interface {
	FindByID(ctx context.Context, ID uuid.UUID) (types.ChatRoom, error)
	CreateTx(ctx context.Context, tx dbUtil.Tx, req types.ChatRoom) error
	FindByUserIDAndServiceID(ctx context.Context, userID, serviceID uuid.UUID) (types.ChatRoom, error)
	FindByUserIDAndOfferID(ctx context.Context, userID, offerID uuid.UUID) (types.ChatRoom, error)
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
			offer_id,
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

func (r *chatRoomImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.ChatRoom) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO chat_rooms (
			id,
			service_id,
			offer_id,
			created_at
		)
		VALUES (
			:id, 
			:service_id, 
			:offer_id,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *chatRoomImpl) FindByUserIDAndServiceID(ctx context.Context, userID, serviceID uuid.UUID) (types.ChatRoom, error) {
	res := types.ChatRoom{}

	query := `
		SELECT
			chat_rooms.id,
			chat_rooms.service_id,
			chat_rooms.offer_id,
			chat_rooms.created_at
		FROM chat_rooms
		INNER JOIN chat_room_users 
			ON chat_room_users.chat_room_id = chat_rooms.id
		WHERE 
			chat_room_users.user_id = $1 
			AND chat_rooms.service_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, userID, serviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *chatRoomImpl) FindByUserIDAndOfferID(ctx context.Context, userID, offerID uuid.UUID) (types.ChatRoom, error) {
	res := types.ChatRoom{}

	query := `
		SELECT
			chat_rooms.id,
			chat_rooms.service_id,
			chat_rooms.offer_id,
			chat_rooms.created_at
		FROM chat_rooms
		INNER JOIN chat_room_users 
			ON chat_room_users.chat_room_id = chat_rooms.id
		WHERE 
			chat_room_users.user_id = $1 
			AND chat_rooms.offer_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, userID, offerID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
