package types

import (
	"github.com/google/uuid"
	"github.com/volatiletech/null/v9"
)

// region repo types

type ServiceProviderArea struct {
	ServiceProviderID uuid.UUID  `db:"service_provider_id"`
	ProvinceID        int64      `db:"province_id"`
	CityID            int64      `db:"city_id"`
	DistrictID        null.Int64 `db:"district_id"`
}

// end of region repo types
