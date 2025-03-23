package types

import (
	"time"

	"github.com/google/uuid"
)

// region repo types

type ChatMessage struct {
	ID          uuid.UUID              `db:"id"`
	ChatRoomID  uuid.UUID              `db:"chat_room_id"`
	UserID      uuid.UUID              `db:"user_id"`
	Content     string                 `db:"content"`
	ContentType ChatMessageContentType `db:"content_type"`
	Read        bool                   `db:"read"`
	CreatedAt   time.Time              `db:"created_at"`
}

type ChatMessageContentType string

const (
	ChatMessageContentTypeText  ChatMessageContentType = "text"
	ChatMessageContentTypeImage ChatMessageContentType = "image"
	ChatMessageContentTypeVideo ChatMessageContentType = "video"
)

type ChatMessageCountUnread struct {
	ChatRoomID uuid.UUID `db:"chat_room_id"`
	Count      int       `db:"count"`
}

// end of region repo types
