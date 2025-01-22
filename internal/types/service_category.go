package types

import (
	"time"

	"github.com/google/uuid"
)

// region repo types

type ServiceCategory struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// end of region repo types
