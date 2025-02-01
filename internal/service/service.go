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
	Update(ctx context.Context, req types.ServiceUpdateReq) error
}

type serviceImpl struct {
	db                         *sqlx.DB
	serviceIndexRepo           repository.ServiceIndex
	serviceProviderRepo        repository.ServiceProvider
	serviceRepo                repository.Service
	serviceCategoryRepo        repository.ServiceCategory
	serviceServiceCategoryRepo repository.ServiceServiceCategory
	fileSvc                    File
}

func NewService(db *sqlx.DB, serviceIndexRepo repository.ServiceIndex, serviceProviderRepo repository.ServiceProvider, serviceRepo repository.Service, serviceCategoryRepo repository.ServiceCategory, serviceServiceCategoryRepo repository.ServiceServiceCategory, fileSvc File) Service {
	return &serviceImpl{
		db:                         db,
		serviceIndexRepo:           serviceIndexRepo,
		serviceProviderRepo:        serviceProviderRepo,
		serviceRepo:                serviceRepo,
		serviceCategoryRepo:        serviceCategoryRepo,
		serviceServiceCategoryRepo: serviceServiceCategoryRepo,
		fileSvc:                    fileSvc,
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

	tempFiles := []types.TempFile{}
	for _, img := range req.Images {
		file, err := s.fileSvc.GetTemp(ctx, img)
		if err != nil {
			return err
		}

		tempFiles = append(tempFiles, types.TempFile(file))
	}

	imagesPath, err := s.fileSvc.BulkUploadToS3(ctx, tempFiles, types.ServiceImageDir)
	if err != nil {
		return err
	}

	service.Images = imagesPath

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
		Rules:           service.Rules,
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

func (s *serviceImpl) Update(ctx context.Context, req types.ServiceUpdateReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(fmt.Sprintf("service provider not found for user_id: %s", req.AuthUser.ID))
	} else if err != nil {
		return err
	}

	service, err := s.serviceRepo.FindByIDAndServiceProviderID(ctx, req.ID, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound})
	} else if err != nil {
		return err
	}

	idxService, err := s.serviceIndexRepo.FindByID(ctx, service.ID.String())
	if errors.Is(err, types.ErrNoData) {
		return errors.New(fmt.Sprintf("service index not found for service_id: %s", service.ID))
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

	serviceCategories := []types.ServiceServiceCategory{}
	idxCategories := []string{}
	for _, category := range categories {
		serviceCategories = append(serviceCategories, types.ServiceServiceCategory{
			ServiceID:         service.ID,
			ServiceCategoryID: category.ID,
		})
		idxCategories = append(idxCategories, category.Name)
	}

	service.Name = req.Name
	service.Description = req.Description
	service.DeliveryMethods = req.DeliveryMethods
	service.FeeStartAt = req.FeeStartAt
	service.FeeEndAt = req.FeeEndAt
	service.Rules = req.Rules
	service.IsAvailable = req.IsAvailable

	idxService.Name = service.Name
	idxService.Description = service.Description
	idxService.DeliveryMethods = service.DeliveryMethods
	idxService.Categories = idxCategories
	idxService.Rules = service.Rules
	idxService.FeeStartAt = service.FeeStartAt
	idxService.FeeEndAt = service.FeeEndAt
	idxService.IsAvailable = service.IsAvailable

	tx, err := dbUtil.NewSqlxTx(ctx, s.db, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	if err := s.serviceServiceCategoryRepo.DeleteByServiceIDTx(ctx, tx, service.ID); err != nil {
		return err
	}

	if err := s.serviceServiceCategoryRepo.BulkCreateTx(ctx, tx, serviceCategories); err != nil {
		return err
	}

	if err := s.serviceRepo.UpdateTx(ctx, tx, service); err != nil {
		return err
	}

	if err := s.serviceIndexRepo.Update(ctx, idxService); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.New(err)
	}

	return nil
}
