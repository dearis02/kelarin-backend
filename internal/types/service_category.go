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

type ServiceCategoryWithServiceID struct {
	ID        uuid.UUID `db:"id"`
	ServiceID uuid.UUID `db:"service_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// end of region repo types

// region service types

type ServiceCategoryRes struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ServiceCategoryGetAllRes struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// end region of service types
