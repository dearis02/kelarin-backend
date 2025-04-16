package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"net/http"
	"strconv"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type ConsumerService interface {
	GetAll(ctx context.Context, req types.ConsumerServiceGetAllReq) ([]types.ConsumerServiceGetAllRes, types.PaginationRes, types.ConsumerServiceGetAllMetadata, error)
	GetByID(ctx context.Context, ID uuid.UUID) (types.ConsumerServiceGetByIDRes, error)
}

type consumerServiceImpl struct {
	serviceIndexRepo        repository.ServiceIndex
	serviceRepo             repository.Service
	serviceProviderAreaRepo repository.ServiceProviderArea
	serviceProviderRepo     repository.ServiceProvider
	fileSvc                 File
}

func NewConsumerService(serviceIndexRepo repository.ServiceIndex, serviceRepo repository.Service, serviceProviderAreaRepo repository.ServiceProviderArea, serviceProviderRepo repository.ServiceProvider, fileSvc File) ConsumerService {
	return &consumerServiceImpl{
		serviceIndexRepo:        serviceIndexRepo,
		serviceRepo:             serviceRepo,
		serviceProviderAreaRepo: serviceProviderAreaRepo,
		serviceProviderRepo:     serviceProviderRepo,
		fileSvc:                 fileSvc,
	}
}

func (s *consumerServiceImpl) GetAll(ctx context.Context, req types.ConsumerServiceGetAllReq) ([]types.ConsumerServiceGetAllRes, types.PaginationRes, types.ConsumerServiceGetAllMetadata, error) {
	res := []types.ConsumerServiceGetAllRes{}
	paginationRes := types.PaginationRes{}
	metadata := types.ConsumerServiceGetAllMetadata{}

	if err := req.ValidateAndNormalize(); err != nil {
		return res, paginationRes, metadata, err
	}

	sizeInt, err := strconv.Atoi(req.Size)
	if err != nil {
		return res, paginationRes, metadata, errors.New(err)
	}

	filter := types.ServiceIndexFilter{
		Limit:           sizeInt,
		LatestTimestamp: req.LatestTimestamp,
		Province:        req.Province,
		City:            req.City,
		Categories:      req.Categories,
		Keyword:         req.Keyword,
	}

	services, totalItem, latestTimestamp, err := s.serviceIndexRepo.FindAllByFilter(ctx, filter)
	if err != nil {
		return res, paginationRes, metadata, err
	}

	metadata.LatestTimestamp = latestTimestamp

	paginationRes = req.GeneratePaginationResponse(totalItem)

	serviceProviderIDs := []uuid.UUID{}
	for _, service := range services {
		serviceProviderIDs = append(serviceProviderIDs, service.ServiceProviderID)
	}

	serviceProviders, err := s.serviceProviderRepo.FindByIDs(ctx, serviceProviderIDs)
	if err != nil {
		return res, paginationRes, metadata, err
	}

	for _, service := range services {
		serviceProvider, exs := lo.Find(serviceProviders, func(v types.ServiceProvider) bool {
			return v.ID == service.ServiceProviderID
		})

		if !exs {
			return res, paginationRes, metadata, errors.Errorf("service provider not found: id %s", service.ServiceProviderID)
		}

		imgURL, err := s.fileSvc.GetS3PresignedURL(ctx, service.Images[0])
		if err != nil {
			return res, paginationRes, metadata, err
		}

		res = append(res, types.ConsumerServiceGetAllRes{
			ID:                    service.ID,
			Name:                  service.Name,
			ImageURL:              imgURL,
			FeeStartAt:            service.FeeStartAt,
			FeeEndAt:              service.FeeEndAt,
			Address:               serviceProvider.Address,
			Province:              service.Province.String,
			City:                  service.City.String,
			ReceivedRatingCount:   service.ReceivedRatingCount,
			ReceivedRatingAverage: service.ReceivedRatingAverage,
		})
	}

	return res, paginationRes, metadata, nil
}

func (s *consumerServiceImpl) GetByID(ctx context.Context, ID uuid.UUID) (types.ConsumerServiceGetByIDRes, error) {
	res := types.ConsumerServiceGetByIDRes{}

	service, err := s.serviceRepo.FindByID(ctx, ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "service not found"})
	} else if err != nil {
		return res, err
	}

	provider, err := s.serviceProviderRepo.FindByID(ctx, service.ServiceProviderID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("service provider not found: id %s", service.ServiceProviderID)
	} else if err != nil {
		return res, err
	}

	providerArea, err := s.serviceProviderAreaRepo.FindByServiceProviderID(ctx, provider.ID)
	if !errors.Is(err, types.ErrNoData) && err != nil {
		return res, err
	}

	serviceImgURLs := []string{}
	for _, img := range service.Images {
		url, err := s.fileSvc.GetS3PresignedURL(ctx, img)
		if err != nil {
			return res, err
		}

		serviceImgURLs = append(serviceImgURLs, url)
	}

	providerLogoURL, err := s.fileSvc.GetS3PresignedURL(ctx, provider.LogoImage)
	if err != nil {
		return res, err
	}

	res = types.ConsumerServiceGetByIDRes{
		ID:                    service.ID,
		Name:                  service.Name,
		Description:           service.Description,
		DeliveryMethods:       service.DeliveryMethods,
		ImageURLs:             serviceImgURLs,
		FeeStartAt:            service.FeeStartAt,
		FeeEndAt:              service.FeeEndAt,
		Rules:                 service.Rules,
		IsAvailable:           service.IsAvailable,
		ReceivedRatingCount:   service.ReceivedRatingCount,
		ReceivedRatingAverage: service.ReceivedRatingAverage,
		ServiceProvider: types.ConsumerServiceServiceProviderRes{
			ID:                    provider.ID,
			Name:                  provider.Name,
			Description:           provider.Description,
			Province:              providerArea.ProvinceName.String,
			City:                  providerArea.CityName.String,
			Address:               provider.Address,
			MobilePhoneNumber:     provider.MobilePhoneNumber,
			Telephone:             provider.Telephone,
			LogoImageURL:          providerLogoURL,
			ReceivedRatingCount:   provider.ReceivedRatingCount,
			ReceivedRatingAverage: provider.ReceivedRatingAverage,
			JoinedAt:              provider.CreatedAt.Format(time.DateOnly),
		},
	}

	return res, nil
}
