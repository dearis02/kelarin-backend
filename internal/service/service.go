package service

import (
	"context"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	Create(ctx context.Context, req types.ServiceCreateReq) error
	GetByID(ctx context.Context, req types.ServiceGetByIDReq) (types.ServiceGetByIDRes, error)
}

type serviceImpl struct {
	db                         *sqlx.DB
	serviceIndexRepo           repository.ServiceIndex
	serviceProviderRepo        repository.ServiceProvider
	serviceRepo                repository.Service
	serviceCategoryRepo        repository.ServiceCategory
	serviceServiceCategoryRepo repository.ServiceServiceCategory
}

func NewService(db *sqlx.DB, serviceIndexRepo repository.ServiceIndex, serviceProviderRepo repository.ServiceProvider, serviceRepo repository.Service, serviceCategoryRepo repository.ServiceCategory, serviceServiceCategoryRepo repository.ServiceServiceCategory,
) Service {
	return &serviceImpl{
		db:                         db,
		serviceIndexRepo:           serviceIndexRepo,
		serviceProviderRepo:        serviceProviderRepo,
		serviceRepo:                serviceRepo,
		serviceCategoryRepo:        serviceCategoryRepo,
		serviceServiceCategoryRepo: serviceServiceCategoryRepo,
	}
}

func (s *serviceImpl) Create(ctx context.Context, req types.ServiceCreateReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(fmt.Sprintf("service provider not found for user_id: %s", req.AuthUser.ID))
	} else if err != nil {
		return err
	}

	categories, err := s.serviceCategoryRepo.FindByIDs(ctx, req.CategoryIDs)
	if err != nil {
		return err
	}

	if len(categories) != len(req.CategoryIDs) {
		return errors.New(types.AppErr{
			Code:    http.StatusBadRequest,
			Message: "invalid category_ids",
		})
	}

	timeNow := time.Now()

	id, err := uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}

	service := types.Service{
		ID:                id,
		ServiceProviderID: provider.ID,
		Name:              req.Name,
		Description:       req.Description,
		DeliveryMethods:   req.DeliveryMethods,
		FeeStartAt:        req.FeeStartAt,
		FeeEndAt:          req.FeeEndAt,
		Rules:             req.Rules,
		IsAvailable:       req.IsAvailable,
		CreatedAt:         timeNow,
	}

	serviceCategories := []types.ServiceServiceCategory{}
	for _, category := range categories {
		serviceCategories = append(serviceCategories, types.ServiceServiceCategory{
			ServiceID:         service.ID,
			ServiceCategoryID: category.ID,
		})
	}

	tx, err := dbUtil.NewSqlxTx(ctx, s.db, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	if err := s.serviceRepo.CreateTx(ctx, tx, service); err != nil {
		return err
	}

	if err := s.serviceServiceCategoryRepo.BulkCreateTx(ctx, tx, serviceCategories); err != nil {
		return err
	}

	idxCategories := []string{}
	for _, category := range categories {
		idxCategories = append(idxCategories, category.Name)
	}

	indexReq := types.ServiceIndex{
		ID:              service.ID,
		Name:            service.Name,
		Description:     service.Description,
		DeliveryMethods: service.DeliveryMethods,
		Categories:      idxCategories,
		FeeStartAt:      service.FeeStartAt,
		FeeEndAt:        service.FeeEndAt,
		IsAvailable:     service.IsAvailable,
		CreatedAt:       timeNow,
	}

	if err := s.serviceIndexRepo.Index(ctx, indexReq); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.New(err)
	}

	return nil
}

func (s *serviceImpl) GetByID(ctx context.Context, req types.ServiceGetByIDReq) (types.ServiceGetByIDRes, error) {
	res := types.ServiceGetByIDRes{}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(fmt.Sprintf("service provider not found: user_id %s", req.AuthUser.ID))
	} else if err != nil {
		return res, err
	}

	service, err := s.serviceRepo.FindByIDAndServiceProviderID(ctx, req.ID, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound})
	} else if err != nil {
		return res, err
	}

	categories, err := s.serviceCategoryRepo.FindByServiceIDs(ctx, []uuid.UUID{service.ID})
	if err != nil {
		return res, err
	}

	categoryRes := []types.ServiceCategoryRes{}
	for _, category := range categories {
		categoryRes = append(categoryRes, types.ServiceCategoryRes{
			ID:   category.ID,
			Name: category.Name,
		})
	}

	res = types.ServiceGetByIDRes{
		ID:              service.ID,
		Name:            service.Name,
		Description:     service.Description,
		DeliveryMethods: service.DeliveryMethods,
		Categories:      categoryRes,
		FeeStartAt:      service.FeeStartAt,
		FeeEndAt:        service.FeeEndAt,
		Rules:           service.Rules,
		IsAvailable:     service.IsAvailable,
		CreatedAt:       service.CreatedAt,
	}

	return res, nil
}
