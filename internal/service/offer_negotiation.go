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
	"github.com/shopspring/decimal"
)

type OfferNegotiation interface {
	ProviderCreate(ctx context.Context, req types.OfferNegotiationProviderCreateReq) error
	ConsumerAction(ctx context.Context, req types.OfferNegotiationConsumerActionReq) error
}

type offerNegotiationImpl struct {
	serviceProviderRepo      repository.ServiceProvider
	offerNegotiationRepo     repository.OfferNegotiation
	offerRepo                repository.Offer
	serviceRepo              repository.Service
	db                       *sqlx.DB
	notificationSvc          Notification
	fcmTokenRepo             repository.FCMToken
	fileSvc                  File
	consumerNotificationRepo repository.ConsumerNotification
}

func NewOfferNegotiation(serviceProviderRepo repository.ServiceProvider, offerNegotiationRepo repository.OfferNegotiation, offerRepo repository.Offer, serviceRepo repository.Service, db *sqlx.DB, notificationSvc Notification, fcmTokenRepo repository.FCMToken, fileSvc File, consumerNotificationRepo repository.ConsumerNotification) OfferNegotiation {
	return &offerNegotiationImpl{
		serviceProviderRepo:      serviceProviderRepo,
		offerNegotiationRepo:     offerNegotiationRepo,
		offerRepo:                offerRepo,
		serviceRepo:              serviceRepo,
		db:                       db,
		notificationSvc:          notificationSvc,
		fcmTokenRepo:             fcmTokenRepo,
		fileSvc:                  fileSvc,
		consumerNotificationRepo: consumerNotificationRepo,
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

	now := time.Now()

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
		CreatedAt:            now,
	}

	id, err = uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}
	consumerNotification := types.ConsumerNotification{
		ID:                 id,
		UserID:             offer.UserID,
		OfferNegotiationID: uuid.NullUUID{UUID: offerNegotiation.ID, Valid: true},
		Type:               types.ConsumerNotificationTypeOfferNegotiationReceived,
		CreatedAt:          now,
	}

	tx, err := dbUtil.NewSqlxTx(ctx, s.db, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	if err = s.offerNegotiationRepo.CreateTx(ctx, tx, offerNegotiation); err != nil {
		return err
	}

	if err := s.consumerNotificationRepo.CreateTx(ctx, tx, consumerNotification); err != nil {
		return err
	}

	key := types.FCMTokenKey(offer.UserID)
	fcmToken, err := s.fcmTokenRepo.Find(ctx, key)
	if !errors.Is(err, types.ErrNoData) && err != nil {
		return err
	}

	if fcmToken != "" {
		providerLogoURL, err := s.fileSvc.GetS3PresignedURL(ctx, provider.LogoImage)
		if err != nil {
			return err
		}

		err = s.notificationSvc.SendPush(ctx, types.NotificationSendReq{
			Title:    fmt.Sprintf("%s want to negotiate", provider.Name),
			Message:  req.Message,
			Token:    fcmToken,
			ImageURL: providerLogoURL,
		})
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *offerNegotiationImpl) ConsumerAction(ctx context.Context, req types.OfferNegotiationConsumerActionReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	negotiation, err := s.offerNegotiationRepo.FindByIDAndUserID(ctx, req.ID, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound, Message: "offer negotiation not found"})
	} else if err != nil {
		return err
	}

	offer, err := s.offerRepo.FindByIDAndUserID(ctx, negotiation.OfferID, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.Errorf("offer not found: id %s", negotiation.OfferID)
	} else if err != nil {
		return err
	}

	if negotiation.Status != types.OfferNegotiationStatusPending {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "only pending negotiation can be accept or reject"})
	} else if offer.Status != types.OfferStatusPending {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "offer is already accepted, rejected, or canceled"})
	}

	switch req.Action {
	case types.OfferNegotiationConsumerActionAccept:
		negotiation.Status = types.OfferNegotiationStatusAccepted
		offer.ServiceCost = negotiation.RequestedServiceCost
	case types.OfferNegotiationConsumerActionReject:
		negotiation.Status = types.OfferNegotiationStatusRejected
	}

	tx, err := dbUtil.NewSqlxTx(ctx, s.db, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	if err := s.offerNegotiationRepo.UpdateStatusTx(ctx, tx, negotiation); err != nil {
		return err
	}

	if err := s.offerRepo.UpdateTx(ctx, tx, offer); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
