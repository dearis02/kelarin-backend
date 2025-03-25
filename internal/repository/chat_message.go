package repository

import (
	"context"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ChatMessage interface {
	CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.ChatMessage) error
	CountUnreadReceivedByChatRoomIDs(ctx context.Context, userID uuid.UUID, chatRoomIDs uuid.UUIDs) ([]types.ChatMessageCountUnread, error)
	FindByChatRoomID(ctx context.Context, roomID uuid.UUID) ([]types.ChatMessage, error)
	FindLatestByChatRoomIDs(ctx context.Context, roomIDs uuid.UUIDs) ([]types.ChatMessage, error)
}

type chatMessageImpl struct {
	db *sqlx.DB
}

func NewChatMessage(db *sqlx.DB) ChatMessage {
	return &chatMessageImpl{db: db}
}

func (r *chatMessageImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, req types.ChatMessage) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO chat_messages (
			id,
			chat_room_id,
			user_id,
			content,
			content_type,
			created_at
		)
		VALUES (
			:id,
			:chat_room_id,
			:user_id,
			:content,
			:content_type,
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *chatMessageImpl) CountUnreadReceivedByChatRoomIDs(ctx context.Context, userID uuid.UUID, chatRoomIDs uuid.UUIDs) ([]types.ChatMessageCountUnread, error) {
	res := []types.ChatMessageCountUnread{}

	query := `
		SELECT
			chat_room_id,
			COUNT(id) AS count
		FROM chat_messages
		WHERE chat_room_id = ANY($1)
			AND read = false
			AND user_id != $2
		GROUP BY chat_room_id
	`

	if err := r.db.SelectContext(ctx, &res, query, pq.Array(chatRoomIDs), userID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *chatMessageImpl) FindByChatRoomID(ctx context.Context, roomID uuid.UUID) ([]types.ChatMessage, error) {
	res := []types.ChatMessage{}

	query := `
		SELECT
			id,
			chat_room_id,
			user_id,
			content,
			content_type,
			read,
			created_at
		FROM chat_messages
		WHERE chat_room_id = $1
		ORDER BY id ASC
	`

	if err := r.db.SelectContext(ctx, &res, query, roomID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *chatMessageImpl) FindLatestByChatRoomIDs(ctx context.Context, roomIDs uuid.UUIDs) ([]types.ChatMessage, error) {
	res := []types.ChatMessage{}

	query := `
		SELECT DISTINCT ON (chat_room_id)
			id,
			chat_room_id,
			user_id,
			content,
			content_type,
			read,
			created_at
		FROM chat_messages
		WHERE chat_room_id = ANY($1)
		ORDER BY chat_room_id, id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, pq.Array(roomIDs)); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
