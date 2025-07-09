package service

import (
	"context"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"
	"net/http"
	"slices"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/volatiletech/null/v9"
)

type Service interface {
	GetAll(ctx context.Context, req types.ServiceGetAllReq) ([]types.ServiceGetAllRes, error)
	Create(ctx context.Context, req types.ServiceCreateReq) error
	GetByID(ctx context.Context, req types.ServiceGetByIDReq) (types.ServiceGetByIDRes, error)
	Update(ctx context.Context, req types.ServiceUpdateReq) error
	Delete(ctx context.Context, req types.ServiceDeleteReq) error
	AddImages(ctx context.Context, req types.ServiceImageActionReq) error
	RemoveImages(ctx context.Context, req types.ServiceImageActionReq) error
}

type serviceImpl struct {
	beginMainDBTx              dbUtil.SqlxTx
	serviceIndexRepo           repository.ServiceIndex
	serviceProviderRepo        repository.ServiceProvider
	serviceRepo                repository.Service
	serviceCategoryRepo        repository.ServiceCategory
	serviceServiceCategoryRepo repository.ServiceServiceCategory
	serviceProviderAreaRepo    repository.ServiceProviderArea
	fileSvc                    File
}

func NewService(beginMainDBTx dbUtil.SqlxTx, serviceIndexRepo repository.ServiceIndex, serviceProviderRepo repository.ServiceProvider, serviceRepo repository.Service, serviceCategoryRepo repository.ServiceCategory, serviceServiceCategoryRepo repository.ServiceServiceCategory, serviceProviderAreaRepo repository.ServiceProviderArea, fileSvc File) Service {
	return &serviceImpl{
		beginMainDBTx:              beginMainDBTx,
		serviceIndexRepo:           serviceIndexRepo,
		serviceProviderRepo:        serviceProviderRepo,
		serviceRepo:                serviceRepo,
		serviceCategoryRepo:        serviceCategoryRepo,
		serviceServiceCategoryRepo: serviceServiceCategoryRepo,
		serviceProviderAreaRepo:    serviceProviderAreaRepo,
		fileSvc:                    fileSvc,
	}
}

func (s *serviceImpl) GetAll(ctx context.Context, req types.ServiceGetAllReq) ([]types.ServiceGetAllRes, error) {
	res := []types.ServiceGetAllRes{}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(fmt.Sprintf("service provider not found: user_id %s", req.AuthUser.ID))
	} else if err != nil {
		return res, err
	}

	services, err := s.serviceRepo.FindAllByServiceProviderID(ctx, provider.ID)
	if err != nil {
		return res, err
	}

	serviceIDs := []uuid.UUID{}
	for _, service := range services {
		serviceIDs = append(serviceIDs, service.ID)
	}

	categories, err := s.serviceCategoryRepo.FindByServiceIDs(ctx, serviceIDs)
	if err != nil {
		return res, err
	}

	for _, service := range services {
		res = append(res, types.ServiceGetAllRes{
			ID:              service.ID,
			Name:            service.Name,
			Description:     service.Description,
			DeliveryMethods: service.DeliveryMethods,
			FeeStartAt:      service.FeeStartAt,
			FeeEndAt:        service.FeeEndAt,
			Rules:           service.Rules,
			IsAvailable:     service.IsAvailable,
			CreatedAt:       service.CreatedAt,
			Categories: lo.Map(categories, func(category types.ServiceCategoryWithServiceID, _ int) types.ServiceCategoryRes {
				return types.ServiceCategoryRes{
					ID:   category.ID,
					Name: category.Name,
				}
			}),
		})
	}

	return res, nil
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

	categoriesMap := lo.SliceToMap(categories, func(category types.ServiceCategory) (uuid.UUID, types.ServiceCategory) {
		return category.ID, category
	})

	for _, id := range req.CategoryIDs {
		if _, ok := categoriesMap[id]; !ok {
			return errors.New(types.AppErr{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("invalid category_id: %s", id),
			})
		}
	}

	areaExists := true
	area, err := s.serviceProviderAreaRepo.FindByServiceProviderID(ctx, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		areaExists = false
	} else if err != nil {
		return err
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

	tx, err := s.beginMainDBTx(ctx, nil)
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

	categoriesName := []string{}
	for _, category := range categories {
		categoriesName = append(categoriesName, category.Name)
	}

	indexReq := types.ServiceIndex{
		ID:                service.ID,
		ServiceProviderID: service.ServiceProviderID,
		Name:              service.Name,
		Description:       service.Description,
		DeliveryMethods:   service.DeliveryMethods,
		Categories:        categoriesName,
		Rules:             service.Rules,
		FeeStartAt:        service.FeeStartAt,
		FeeEndAt:          service.FeeEndAt,
		IsAvailable:       service.IsAvailable,
		Images:            service.Images,
		CreatedAt:         timeNow,
	}

	if areaExists {
		indexReq.Province = area.ProvinceName
		indexReq.City = area.CityName
	}

	if err := s.serviceIndexRepo.Create(ctx, indexReq); err != nil {
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

	images := []types.ImageRes{}
	for _, img := range service.Images {
		url, err := s.fileSvc.GetS3PresignedURL(ctx, img)
		if err != nil {
			return res, err
		}
		images = append(images, types.ImageRes{
			Key: img,
			URL: url,
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
		Images:          images,
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

	area, err := s.serviceProviderAreaRepo.FindByServiceProviderID(ctx, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		// do noting
	} else if err != nil {
		return err
	}

	idxService, seqNo, primaryTerm, err := s.serviceIndexRepo.FindByID(ctx, service.ID.String())
	if errors.Is(err, types.ErrNoData) {
		return errors.New(fmt.Sprintf("service index not found for service_id: %s", service.ID))
	} else if err != nil {
		return err
	}

	categories, err := s.serviceCategoryRepo.FindByIDs(ctx, req.CategoryIDs)
	if err != nil {
		return err
	}

	categoryIDsMap := lo.SliceToMap(categories, func(category types.ServiceCategory) (uuid.UUID, any) {
		return category.ID, struct{}{}
	})

	for _, id := range req.CategoryIDs {
		if _, ok := categoryIDsMap[id]; !ok {
			return errors.New(types.AppErr{
				Code:    http.StatusUnprocessableEntity,
				Message: fmt.Sprintf("category_ids: id %s is invalid", id),
			})
		}
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

	if area.ID != 0 {
		idxService.Province = area.ProvinceName
		idxService.City = area.CityName
	}

	tx, err := s.beginMainDBTx(ctx, nil)
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

	err = s.serviceRepo.UpdateTx(ctx, tx, service)
	if err != nil {
		return err
	}

	if err := s.serviceIndexRepo.Update(ctx, idxService, seqNo, primaryTerm); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.New(err)
	}

	return nil
}

func (s *serviceImpl) Delete(ctx context.Context, req types.ServiceDeleteReq) error {
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

	idxService, _, _, err := s.serviceIndexRepo.FindByID(ctx, service.ID.String())
	if errors.Is(err, types.ErrNoData) {
		return errors.New(fmt.Sprintf("service index not found for service_id: %s", service.ID))
	} else if err != nil {
		return err
	}

	tx, err := s.beginMainDBTx(ctx, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	service.DeletedAt = null.TimeFrom(time.Now())
	if err = s.serviceRepo.DeleteTx(ctx, tx, service); err != nil {
		return err
	}

	if err = s.serviceIndexRepo.Delete(ctx, idxService); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.New(err)
	}

	return nil
}

func (s *serviceImpl) AddImages(ctx context.Context, req types.ServiceImageActionReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(fmt.Sprintf("service provider not found : user_id: %s", req.AuthUser.ID))
	} else if err != nil {
		return err
	}

	service, err := s.serviceRepo.FindByIDAndServiceProviderID(ctx, req.ID, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound})
	} else if err != nil {
		return err
	}

	tempFiles := []types.TempFile{}
	for _, img := range req.ImageKeys {
		file, err := s.fileSvc.GetTemp(ctx, img)
		if err != nil {
			return err
		}

		tempFiles = append(tempFiles, types.TempFile(file))
	}

	imgKeys, err := s.fileSvc.BulkUploadToS3(ctx, tempFiles, types.ServiceImageDir)
	if err != nil {
		return err
	}

	service.Images = append(service.Images, imgKeys...)

	tx, err := s.beginMainDBTx(ctx, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	if err := s.serviceRepo.UpdateTx(ctx, tx, service); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return errors.New(err)
	}

	return nil
}

func (s *serviceImpl) RemoveImages(ctx context.Context, req types.ServiceImageActionReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(fmt.Sprintf("service provider not found : user_id: %s", req.AuthUser.ID))
	} else if err != nil {
		return err
	}

	service, err := s.serviceRepo.FindByIDAndServiceProviderID(ctx, req.ID, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound})
	} else if err != nil {
		return err
	}

	for _, k := range req.ImageKeys {
		if exs := slices.Contains(service.Images, k); !exs {
			return errors.New(types.AppErr{Code: http.StatusNotFound, Message: fmt.Sprintf("image key not found: %s", k)})
		}
	}

	images := []string{}
	for _, img := range service.Images {
		if !slices.Contains(req.ImageKeys, img) {
			images = append(images, img)
		}
	}

	service.Images = images

	tx, err := s.beginMainDBTx(ctx, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	if err := s.serviceRepo.UpdateTx(ctx, tx, service); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return errors.New(err)
	}

	return nil
}
