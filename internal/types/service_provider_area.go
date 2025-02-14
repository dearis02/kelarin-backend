package types

import (
	"github.com/google/uuid"
	"github.com/volatiletech/null/v9"
)

// region repo types

type ServiceProviderArea struct {
	ID                int64      `db:"id"`
	ServiceProviderID uuid.UUID  `db:"service_provider_id"`
	ProvinceID        int64      `db:"province_id"`
	CityID            int64      `db:"city_id"`
	DistrictID        null.Int64 `db:"district_id"`
}

type ServiceProviderAreaWithAreaDetail struct {
	ID                int64       `db:"id"`
	ServiceProviderID uuid.UUID   `db:"service_provider_id"`
	ProvinceID        int64       `db:"province_id"`
	CityID            int64       `db:"city_id"`
	DistrictID        null.Int64  `db:"district_id"`
	ProvinceName      null.String `db:"province_name"`
	CityName          null.String `db:"city_name"`
}

// end of region repo types
