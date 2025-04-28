package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"kelarin/internal/utils"
	dbUtil "kelarin/internal/utils/dbutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type ConsumerService interface {
	GetAll(ctx context.Context, req types.ConsumerServiceGetAllReq) ([]types.ConsumerServiceGetAllRes, types.PaginationRes, error)
	GetByID(ctx context.Context, ID uuid.UUID) (types.ConsumerServiceGetByIDRes, error)
	CreateFeedback(ctx context.Context, req types.ConsumerServiceFeedbackCreateReq) error
}

type consumerServiceImpl struct {
	beginMainDBTx           dbUtil.SqlxTx
	serviceIndexRepo        repository.ServiceIndex
	serviceRepo             repository.Service
	serviceProviderAreaRepo repository.ServiceProviderArea
	serviceProviderRepo     repository.ServiceProvider
	fileSvc                 File
	orderRepo               repository.Order
	serviceFeedbackRepo     repository.ServiceFeedback
}

func NewConsumerService(
	beginMainDBTx dbUtil.SqlxTx,
	serviceIndexRepo repository.ServiceIndex,
	serviceRepo repository.Service,
	serviceProviderAreaRepo repository.ServiceProviderArea,
	serviceProviderRepo repository.ServiceProvider,
	fileSvc File,
	orderRepo repository.Order,
	serviceFeedbackRepo repository.ServiceFeedback,
) ConsumerService {
	return &consumerServiceImpl{
		beginMainDBTx:           beginMainDBTx,
		serviceIndexRepo:        serviceIndexRepo,
		serviceRepo:             serviceRepo,
		serviceProviderAreaRepo: serviceProviderAreaRepo,
		serviceProviderRepo:     serviceProviderRepo,
		fileSvc:                 fileSvc,
		orderRepo:               orderRepo,
		serviceFeedbackRepo:     serviceFeedbackRepo,
	}
}

func (s *consumerServiceImpl) GetAll(ctx context.Context, req types.ConsumerServiceGetAllReq) ([]types.ConsumerServiceGetAllRes, types.PaginationRes, error) {
	res := []types.ConsumerServiceGetAllRes{}
	paginationRes := types.PaginationRes{}

	if err := req.ValidateAndNormalize(); err != nil {
		return res, paginationRes, err
	}

	sizeInt, err := strconv.Atoi(req.Size)
	if err != nil {
		return res, paginationRes, errors.New(err)
	}

	after, err := utils.DecodeESAfter(req.After)
	if err != nil {
		return res, paginationRes, err
	}

	filter := types.ServiceIndexFilter{
		Limit:      sizeInt,
		After:      after,
		Province:   req.Province,
		City:       req.City,
		Categories: req.Categories,
		Keyword:    req.Keyword,
	}

	services, totalItem, after, err := s.serviceIndexRepo.FindAllByFilter(ctx, filter)
	if err != nil {
		return res, paginationRes, err
	}

	afterRes, err := utils.EncodeEsAfter(after)
	if err != nil {
		return res, paginationRes, err
	}

	paginationRes = req.GeneratePaginationResponse(totalItem)
	paginationRes.After = afterRes

	serviceProviderIDs := []uuid.UUID{}
	for _, service := range services {
		serviceProviderIDs = append(serviceProviderIDs, service.ServiceProviderID)
	}

	serviceProviders, err := s.serviceProviderRepo.FindByIDs(ctx, serviceProviderIDs)
	if err != nil {
		return res, paginationRes, err
	}

	for _, service := range services {
		serviceProvider, exs := lo.Find(serviceProviders, func(v types.ServiceProvider) bool {
			return v.ID == service.ServiceProviderID
		})

		if !exs {
			return res, paginationRes, errors.Errorf("service provider not found: id %s", service.ServiceProviderID)
		}

		imgURL, err := s.fileSvc.GetS3PresignedURL(ctx, service.Images[0])
		if err != nil {
			return res, paginationRes, err
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

	return res, paginationRes, nil
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

func (s *consumerServiceImpl) CreateFeedback(ctx context.Context, req types.ConsumerServiceFeedbackCreateReq) error {
	err := req.Validate()
	if err != nil {
		return err
	}

	order, err := s.orderRepo.FindByIDAndUserID(ctx, req.OrderID, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "order not found"})
	} else if err != nil {
		return err
	}

	if order.Status != types.OrderStatusFinished {
		return errors.New(types.AppErr{Code: http.StatusConflict, Message: "order not finished"})
	}

	feedbackGiven := true
	_, err = s.serviceFeedbackRepo.FindByOrderID(ctx, order.ID)
	if errors.Is(err, types.ErrNoData) {
		feedbackGiven = false
	} else if err != nil {
		return err
	}

	if feedbackGiven {
		return errors.New(types.AppErr{Code: http.StatusConflict, Message: "feedback already given"})
	}

	timeNow := time.Now()

	id, err := uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}

	feedback := types.ServiceFeedback{
		ID:        id,
		ServiceID: order.ServiceID,
		OrderID:   order.ID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: timeNow,
	}

	idxService, seqNo, primaryTerm, err := s.serviceIndexRepo.FindByID(ctx, order.ServiceID.String())
	if errors.Is(err, types.ErrNoData) {
		return errors.Errorf("service index not found: id %s", order.ServiceID)
	} else if err != nil {
		return err
	}

	tx, err := s.beginMainDBTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = s.serviceFeedbackRepo.CreateTx(ctx, tx, feedback)
	if err != nil {
		return err
	}

	recvRatingCount, recvRatingAverage, err := s.serviceRepo.UpdateAsFeedbackGiven(ctx, tx, order.ServiceID, req.Rating)
	if err != nil {
		return err
	}

	idxService.ReceivedRatingCount = recvRatingCount
	idxService.ReceivedRatingAverage = recvRatingAverage

	err = s.serviceIndexRepo.Update(ctx, idxService, seqNo, primaryTerm)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return errors.New(err)
	}

	return nil
}
