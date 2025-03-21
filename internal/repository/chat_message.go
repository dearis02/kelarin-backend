package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ChatMessage interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.ChatMessage) error
	FindAllUnreadReceivedByChatRoomIDs(ctx context.Context, userID uuid.UUID, chatRoomIDs uuid.UUIDs) ([]types.ChatMessage, error)
	FindByChatRoomID(ctx context.Context, roomID uuid.UUID) ([]types.ChatMessage, error)
}

type chatMessageImpl struct {
	db *sqlx.DB
}

func NewChatMessage(db *sqlx.DB) ChatMessage {
	return &chatMessageImpl{db: db}
}

func (r *chatMessageImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, req types.ChatMessage) error {
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

func (r *chatMessageImpl) FindAllUnreadReceivedByChatRoomIDs(ctx context.Context, userID uuid.UUID, chatRoomIDs uuid.UUIDs) ([]types.ChatMessage, error) {
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
		WHERE chat_room_id = ANY($1)
			AND read = false
			AND user_id != $2
		ORDER BY id DESC
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
		ORDER BY id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, roomID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
