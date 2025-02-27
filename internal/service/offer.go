package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Offer interface {
	ConsumerCreate(ctx context.Context, req types.OfferConsumerCreateReq) error
	ConsumerGetAll(ctx context.Context, req types.OfferConsumerGetAllReq) ([]types.OfferConsumerGetAllRes, error)
}

type offerImpl struct {
	offerRepo       repository.Offer
	userAddressRepo repository.UserAddress
	serviceRepo     repository.Service
	fileSvc         File
}

func NewOffer(offerRepo repository.Offer, userAddressRepo repository.UserAddress, serviceRepo repository.Service, fileSvc File) Offer {
	return &offerImpl{
		offerRepo:       offerRepo,
		userAddressRepo: userAddressRepo,
		serviceRepo:     serviceRepo,
		fileSvc:         fileSvc,
	}
}

func (s *offerImpl) ConsumerCreate(ctx context.Context, req types.OfferConsumerCreateReq) error {
	service, err := s.serviceRepo.FindByID(ctx, req.ServiceID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound, Message: "offer not found"})
	} else if err != nil {
		return err
	}

	if err := req.Validate(service.FeeStartAt); err != nil {
		return err
	}

	exs, err := s.offerRepo.IsPendingOfferExists(ctx, req.AuthUser.ID, service.ID)
	if err != nil {
		return errors.New(err)
	}

	if exs {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "there is still pending offer for this service"})
	}

	address, err := s.userAddressRepo.FindByIDAndUserID(ctx, req.AddressID, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound, Message: "address not found"})
	} else if err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}

	startDate, err := time.Parse(time.DateOnly, req.ServiceStartDate)
	if err != nil {
		return errors.New(err)
	}

	endDate, err := time.Parse(time.DateOnly, req.ServiceEndDate)
	if err != nil {
		return errors.New(err)
	}

	startTime, err := time.Parse(time.TimeOnly, req.ServiceStartTime)
	if err != nil {
		return err
	}

	endTime, err := time.Parse(time.TimeOnly, req.ServiceEndTime)
	if err != nil {
		return err
	}

	offer := types.Offer{
		ID:               id,
		UserID:           req.AuthUser.ID,
		UserAddressID:    address.ID,
		ServiceID:        service.ID,
		Detail:           req.Detail,
		ServiceCost:      decimal.NewFromFloat(req.ServiceCost),
		ServiceStartDate: startDate,
		ServiceEndDate:   endDate,
		ServiceStartTime: startTime,
		ServiceEndTime:   endTime,
		Status:           types.OfferStatusPending,
		CreatedAt:        time.Now(),
	}

	if err = s.offerRepo.Create(ctx, offer); err != nil {
		return err
	}

	return nil
}

func (s *offerImpl) ConsumerGetAll(ctx context.Context, req types.OfferConsumerGetAllReq) ([]types.OfferConsumerGetAllRes, error) {
	res := []types.OfferConsumerGetAllRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	offers, err := s.offerRepo.FindAllByUserID(ctx, req.AuthUser.ID)
	if err != nil {
		return res, err
	}

	var reqTimeZone *time.Location
	if req.TimeZone != "" {
		reqTimeZone, err = time.LoadLocation(req.TimeZone)
	} else {
		reqTimeZone, err = time.LoadLocation(types.AppTimeZone)
	}

	if err != nil {
		return res, errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "invalid timezone"})
	}

	for _, o := range offers {
		serviceImgURL, err := s.fileSvc.GetS3PresignedURL(ctx, o.ServiceImage)
		if err != nil {
			return res, err
		}

		serviceProviderLogoURL, err := s.fileSvc.GetS3PresignedURL(ctx, o.ServiceProviderLogo)
		if err != nil {
			return res, err
		}

		res = append(res, types.OfferConsumerGetAllRes{
			ID:                  o.ID,
			ServiceCost:         o.ServiceCost,
			ServiceStartDate:    o.ServiceStartDate.Format(time.DateOnly),
			ServiceEndDate:      o.ServiceEndDate.Format(time.DateOnly),
			ServiceStartTime:    o.ServiceStartTime.In(reqTimeZone).Format(time.TimeOnly),
			ServiceEndTime:      o.ServiceEndTime.In(reqTimeZone).Format(time.TimeOnly),
			ServiceTimeTimeZone: reqTimeZone.String(),
			// HasPendingNegotiation: false, TODO: need to implement this
			CreatedAt: o.CreatedAt,
			Service: types.OfferConsumerGetAllResService{
				ID:       o.ServiceID,
				Name:     o.ServiceName,
				ImageURL: serviceImgURL,
			},
			ServiceProvider: types.OfferConsumerGetAllResServiceProvider{
				ID:      o.ServiceProviderID,
				Name:    o.ServiceProviderName,
				LogoURL: serviceProviderLogoURL,
			},
		})
	}

	return res, nil
}
