package types

import (
	"github.com/alexliesenfeld/opencage"
	"github.com/golang/geo/s2"
)

// region service types

type GeocodingReverseReq struct {
	LatLong s2.LatLng
}

type GeocodingReverseRes opencage.Response

// end of region service types
