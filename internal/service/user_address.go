package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"

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
		Name:     req.Name,
		UserID:   req.AuthUser.ID,
		Province: req.Province,
		City:     req.City,
		Address:  req.Address,
	}

	if req.Lat.Valid && req.Lng.Valid {
		userAddress.Coordinates = null.StringFrom(fmt.Sprintf("POINT(%s %s)", req.Lng.Decimal, req.Lat.Decimal))
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

			lng = null.Float64From(ewkbPoint.X())
			lat = null.Float64From(ewkbPoint.Y())
		}

		res = append(res, types.UserAddressGetAllRes{
			ID:       a.ID,
			Name:     a.Name,
			Lat:      lat,
			Lng:      lng,
			Province: a.Province,
			City:     a.City,
			Address:  a.Address,
		})
	}

	return res, nil
}
