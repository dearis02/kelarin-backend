package types

import "github.com/google/uuid"

// region repo types

type ServiceServiceCategory struct {
	ServiceID         uuid.UUID `db:"service_id"`
	ServiceCategoryID uuid.UUID `db:"service_category_id"`
}

// end of region repo types
