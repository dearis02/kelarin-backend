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

type ServiceCategoryRes struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// end of region repo types
