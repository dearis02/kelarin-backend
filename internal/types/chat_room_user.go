package types

import (
	"time"

	"github.com/google/uuid"
)

// region repo types

type ChatRoomUser struct {
	ChatRoomID uuid.UUID `db:"chat_room_id"`
	UserID     uuid.UUID `db:"user_id"`
	CreatedAt  time.Time `db:"created_at"`
}

// end of region repo types
