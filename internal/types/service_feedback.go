package types

import (
	"time"

	"github.com/google/uuid"
)

// region repo types

type ServiceFeedback struct {
	ID        uuid.UUID `db:"id"`
	ServiceID uuid.UUID `db:"service_id"`
	OrderID   uuid.UUID `db:"order_id"`
	Rating    int16     `db:"rating"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
}

// endregion repo types

// region service types

// endregion service types
