package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"github.com/volatiletech/null/v9"
)

type UserAddress interface {
	Create(ctx context.Context, req types.UserAddressCreateReq) error
	GetAll(ctx context.Context, req types.UserAddressGetAllReq) ([]types.UserAddressGetAllRes, error)
}

type userAddressImpl struct {
	userAddressRepo repository.UserAddress
	geocodingSvc    Geocoding
}

func NewUserAddress(userAddressRepo repository.UserAddress, geocodingSvc Geocoding) UserAddress {
	return &userAddressImpl{
		userAddressRepo: userAddressRepo,
		geocodingSvc:    geocodingSvc,
	}
}

func (s *userAddressImpl) Create(ctx context.Context, req types.UserAddressCreateReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	userAddress := types.UserAddress{
		ID:       id,
		UserID:   req.AuthUser.ID,
		Province: req.Province,
		City:     req.City,
		Address:  req.Address,
	}

	if req.Lat.Valid && req.Lng.Valid {
		lat := req.Lat.Decimal.InexactFloat64()
		lng := req.Lng.Decimal.InexactFloat64()

		reverseGeocodingRes, err := s.geocodingSvc.Reverse(ctx, types.GeocodingReverseReq{
			LatLong: s2.LatLngFromDegrees(lat, lng),
		})
		if err != nil {
			return err
		}

		userAddress.Coordinates = null.StringFrom(fmt.Sprintf("POINT(%f %f)", lat, lng))
		userAddress.Address = reverseGeocodingRes.Results[0].Formatted

		if reverseGeocodingRes.Results[0].Components.City != "" {
			userAddress.City = reverseGeocodingRes.Results[0].Components.City
		} else if reverseGeocodingRes.Results[0].Components.County != "" {
			userAddress.City = reverseGeocodingRes.Results[0].Components.County
		}

		userAddress.Province = reverseGeocodingRes.Results[0].Components.State
	}

	if err = s.userAddressRepo.Create(ctx, userAddress); err != nil {
		return err
	}

	return nil
}

func (s *userAddressImpl) GetAll(ctx context.Context, req types.UserAddressGetAllReq) ([]types.UserAddressGetAllRes, error) {
	res := []types.UserAddressGetAllRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	addresses, err := s.userAddressRepo.FindByUserID(ctx, req.AuthUser.ID)
	if err != nil {
		return res, err
	}

	for _, a := range addresses {
		var lat null.Float64
		var lng null.Float64

		if a.Coordinates.Valid {
			bytes, err := hex.DecodeString(a.Coordinates.String)
			if err != nil {
				return res, err
			}

			var ewkbPoint ewkb.Point
			if err := ewkbPoint.Scan(bytes); err != nil {
				return res, err
			}

			lat = null.Float64From(ewkbPoint.X())
			lng = null.Float64From(ewkbPoint.Y())
		}

		res = append(res, types.UserAddressGetAllRes{
			ID:       a.ID,
			Lat:      lat,
			Lng:      lng,
			Province: a.Province,
			City:     a.City,
			Address:  a.Address,
		})
	}

	return res, nil
}
