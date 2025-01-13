package service

import (
	"context"
	"kelarin/internal/types"

	"github.com/alexliesenfeld/opencage"
)

type Geocoding interface {
	Reverse(ctx context.Context, req types.GeocodingReverseReq) (types.GeocodingReverseRes, error)
}

type geocodingImpl struct {
	openCageClient *opencage.Client
}

func NewGeocoding(openCageClient *opencage.Client) Geocoding {
	return &geocodingImpl{openCageClient}
}

func (s *geocodingImpl) Reverse(ctx context.Context, req types.GeocodingReverseReq) (types.GeocodingReverseRes, error) {
	res := types.GeocodingReverseRes{}

	query := req.LatLong.String()
	params := &opencage.GeocodingParams{
		NoAnnotations: true,
		Language:      "id-ID",
		RoadInfo:      true,
	}

	geocodeRes, err := s.openCageClient.Geocode(ctx, query, params)
	if err != nil {
		return res, err
	}

	return types.GeocodingReverseRes(geocodeRes), nil
}
