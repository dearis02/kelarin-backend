package types

import (
	"time"

	"github.com/google/uuid"
)

// region repo types

type ChatRoom struct {
	ID        uuid.UUID     `db:"id"`
	ServiceID uuid.NullUUID `db:"service_id"`
	CreatedAt time.Time     `db:"created_at"`
}

// end of region repo types
