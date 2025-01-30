package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/jmoiron/sqlx"
)

type ChatMessage interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, req types.ChatMessage) error
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
		return err
	}

	return nil
}
