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

type OfferNegotiation interface {
	ProviderCreate(ctx context.Context, req types.OfferNegotiationProviderCreateReq) error
}

type offerNegotiationImpl struct {
	serviceProviderRepo  repository.ServiceProvider
	offerNegotiationRepo repository.OfferNegotiation
	offerRepo            repository.Offer
	serviceRepo          repository.Service
}

func NewOfferNegotiation(serviceProviderRepo repository.ServiceProvider, offerNegotiationRepo repository.OfferNegotiation, offerRepo repository.Offer, serviceRepo repository.Service) OfferNegotiation {
	return &offerNegotiationImpl{
		serviceProviderRepo:  serviceProviderRepo,
		offerNegotiationRepo: offerNegotiationRepo,
		offerRepo:            offerRepo,
		serviceRepo:          serviceRepo,
	}
}

func (s *offerNegotiationImpl) ProviderCreate(ctx context.Context, req types.OfferNegotiationProviderCreateReq) error {
	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.Errorf("service provider not found: user_id %s", req.AuthUser.ID)
	} else if err != nil {
		return err
	}

	offer, err := s.offerRepo.FindByIDAndServiceProviderID(ctx, req.OfferID, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound, Message: "offer not found"})
	} else if err != nil {
		return err
	}

	negotiation, err := s.offerNegotiationRepo.FindByOfferIDAndStatus(ctx, offer.ID, types.OfferNegotiationStatusPending)
	if !errors.Is(err, types.ErrNoData) && err != nil {
		return err
	}

	if negotiation.Status != "" {
		return errors.New(types.AppErr{
			Code:    http.StatusForbidden,
			Message: "there is still pending negotiation for this offer",
		})
	}

	service, err := s.serviceRepo.FindByID(ctx, offer.ServiceID)
	if errors.Is(err, types.ErrNoData) {
		return errors.Errorf("service not found: id %s", offer.ServiceID)
	} else if err != nil {
		return err
	}

	if err = req.Validate(service.FeeStartAt); err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}

	offerNegotiation := types.OfferNegotiation{
		ID:                   id,
		OfferID:              offer.ID,
		Message:              req.Message,
		RequestedServiceCost: decimal.NewFromFloat(req.RequestedServiceCost),
		Status:               types.OfferNegotiationStatusPending,
		CreatedAt:            time.Now(),
	}

	if err = s.offerNegotiationRepo.Create(ctx, offerNegotiation); err != nil {
		return err
	}

	return nil
}
