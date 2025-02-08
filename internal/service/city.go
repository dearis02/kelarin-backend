package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
)

type City interface {
	GetByProvinceID(ctx context.Context, req types.CityGetByProvinceIDReq) ([]types.CityGetAllRes, error)
}

type cityImpl struct {
	cityRepo repository.City
}

func NewCity(cityRepo repository.City) City {
	return &cityImpl{
		cityRepo: cityRepo,
	}
}

func (s *cityImpl) GetByProvinceID(ctx context.Context, req types.CityGetByProvinceIDReq) ([]types.CityGetAllRes, error) {
	res := []types.CityGetAllRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	cities, err := s.cityRepo.FindByProvinceID(ctx, req.ProvinceID)
	if err != nil {
		return res, err
	}

	for _, c := range cities {
		res = append(res, types.CityGetAllRes(c))
	}

	return res, nil
}
