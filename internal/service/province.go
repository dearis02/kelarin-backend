package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
)

type Province interface {
	GetAll(ctx context.Context) ([]types.ProvinceGetAllRes, error)
}

type provinceImpl struct {
	provinceRepo repository.Province
}

func NewProvince(provinceRepo repository.Province) Province {
	return &provinceImpl{
		provinceRepo: provinceRepo,
	}
}

func (s *provinceImpl) GetAll(ctx context.Context) ([]types.ProvinceGetAllRes, error) {
	res := []types.ProvinceGetAllRes{}

	provinces, err := s.provinceRepo.FindAll(ctx)
	if err != nil {
		return res, err
	}

	for _, p := range provinces {
		res = append(res, types.ProvinceGetAllRes(p))
	}

	return res, nil
}
