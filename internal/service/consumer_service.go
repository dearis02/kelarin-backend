package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"strconv"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
)

type ConsumerService interface {
	GetAll(ctx context.Context, req types.ConsumerServiceGetAllReq) ([]types.ConsumerServiceGetAllRes, types.PaginationRes, types.ConsumerServiceGetAllMetadata, error)
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

	idxServices, totalItem, latestTimestamp, err := s.serviceIndexRepo.FindAllByFilter(ctx, filter)
	if err != nil {
		return res, paginationRes, metadata, err
	}

	metadata.LatestTimestamp = latestTimestamp

	paginationRes = req.GeneratePaginationResponse(totalItem)

	serviceIDs := []uuid.UUID{}
	for _, service := range idxServices {
		serviceIDs = append(serviceIDs, service.ID)
	}

	services, err := s.serviceRepo.FindByIDs(ctx, serviceIDs)
	if err != nil {
		return res, paginationRes, metadata, err
	}

	serviceProviderIDs := []uuid.UUID{}
	for _, service := range services {
		serviceProviderIDs = append(serviceProviderIDs, service.ServiceProviderID)
	}

	serviceProviders, err := s.serviceProviderRepo.FindByIDs(ctx, serviceProviderIDs)
	if err != nil {
		return res, paginationRes, metadata, err
	}

	areas, err := s.serviceProviderAreaRepo.FindByServiceProviderIDs(ctx, serviceProviderIDs)
	if err != nil {
		return res, paginationRes, metadata, err
	}

	for _, service := range services {
		area := types.ServiceProviderAreaWithAreaDetail{}
	areaLoop:
		for _, a := range areas {
			if a.ServiceProviderID == service.ServiceProviderID {
				area = a
				break areaLoop
			}
		}

		serviceProvider := types.ServiceProvider{}
	svcProviderLoop:
		for _, s := range serviceProviders {
			if s.ID == service.ServiceProviderID {
				serviceProvider = s
				break svcProviderLoop
			}
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
			Province:              area.ProvinceName.String,
			City:                  area.CityName.String,
			ReceivedRatingCount:   service.ReceivedRatingCount,
			ReceivedRatingAverage: service.ReceivedRatingAverage,
		})
	}

	return res, paginationRes, metadata, nil
}
