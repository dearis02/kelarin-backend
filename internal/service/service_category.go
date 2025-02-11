package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
)

type ServiceCategory interface {
	GetAll(ctx context.Context) ([]types.ServiceCategoryGetAllRes, error)
}

type serviceCategoryImpl struct {
	serviceCategoryRepo repository.ServiceCategory
}

func NewServiceCategory(serviceCategoryRepo repository.ServiceCategory) ServiceCategory {
	return &serviceCategoryImpl{serviceCategoryRepo: serviceCategoryRepo}
}

func (s *serviceCategoryImpl) GetAll(ctx context.Context) ([]types.ServiceCategoryGetAllRes, error) {
	res := []types.ServiceCategoryGetAllRes{}

	categories, err := s.serviceCategoryRepo.FindAll(ctx)
	if err != nil {
		return res, err
	}

	for _, c := range categories {
		res = append(res, types.ServiceCategoryGetAllRes{
			ID:   c.ID,
			Name: c.Name,
		})
	}

	return res, nil
}
