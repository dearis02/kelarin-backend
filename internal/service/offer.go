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
}

type offerImpl struct {
	offerRepo       repository.Offer
	userAddressRepo repository.UserAddress
	serviceRepo     repository.Service
}

func NewOffer(offerRepo repository.Offer, userAddressRepo repository.UserAddress, serviceRepo repository.Service) Offer {
	return &offerImpl{
		offerRepo:       offerRepo,
		userAddressRepo: userAddressRepo,
		serviceRepo:     serviceRepo,
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

	offer := types.Offer{
		ID:               id,
		UserID:           req.AuthUser.ID,
		UserAddressID:    address.ID,
		ServiceID:        service.ID,
		Detail:           req.Detail,
		ServiceCost:      decimal.NewFromFloat(req.ServiceCost),
		ServiceStartDate: startDate,
		ServiceEndDate:   endDate,
		ServiceStartTime: req.ServiceStartTime,
		ServiceEndTime:   req.ServiceEndTime,
		Status:           types.OfferStatusPending,
		CreatedAt:        time.Now(),
	}

	if err = s.offerRepo.Create(ctx, offer); err != nil {
		return err
	}

	return nil
}
