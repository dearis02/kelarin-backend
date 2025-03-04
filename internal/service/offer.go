package service

import (
	"context"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"kelarin/internal/utils"
	dbUtil "kelarin/internal/utils/dbutil"
	"net/http"
	"slices"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

type Offer interface {
	ConsumerCreate(ctx context.Context, req types.OfferConsumerCreateReq) error
	ConsumerGetAll(ctx context.Context, req types.OfferConsumerGetAllReq) ([]types.OfferConsumerGetAllRes, error)
	ConsumerGetByID(ctx context.Context, req types.OfferConsumerGetByIDReq) (types.OfferConsumerGetByIDRes, error)

	ProviderAction(ctx context.Context, req types.OfferProviderActionReq) error
}

type offerImpl struct {
	offerRepo                       repository.Offer
	userAddressRepo                 repository.UserAddress
	serviceRepo                     repository.Service
	fileSvc                         File
	serviceProviderRepo             repository.ServiceProvider
	offerNegotiationRepo            repository.OfferNegotiation
	serviceProviderNotificationRepo repository.ServiceProviderNotification
	fcmTokenRepo                    repository.FCMToken
	notificationSvc                 Notification
	userRepo                        repository.User
	db                              *sqlx.DB
	consumerNotificationRepo        repository.ConsumerNotification
	chatSvc                         Chat
}

func NewOffer(offerRepo repository.Offer, userAddressRepo repository.UserAddress, serviceRepo repository.Service, fileSvc File, serviceProviderRepo repository.ServiceProvider, offerNegotiationRepo repository.OfferNegotiation, serviceProviderNotificationRepo repository.ServiceProviderNotification, fcmTokenRepo repository.FCMToken, notificationSvc Notification, userRepo repository.User, db *sqlx.DB, consumerNotificationRepo repository.ConsumerNotification, chatSvc Chat) Offer {
	return &offerImpl{
		offerRepo:                       offerRepo,
		userAddressRepo:                 userAddressRepo,
		serviceRepo:                     serviceRepo,
		fileSvc:                         fileSvc,
		serviceProviderRepo:             serviceProviderRepo,
		offerNegotiationRepo:            offerNegotiationRepo,
		serviceProviderNotificationRepo: serviceProviderNotificationRepo,
		fcmTokenRepo:                    fcmTokenRepo,
		notificationSvc:                 notificationSvc,
		userRepo:                        userRepo,
		db:                              db,
		consumerNotificationRepo:        consumerNotificationRepo,
		chatSvc:                         chatSvc,
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

	user, err := s.userRepo.FindByID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.Errorf("user not found: id %s", req.AuthUser.ID)
	} else if err != nil {
		return err
	}

	provider, err := s.serviceProviderRepo.FindByID(ctx, service.ServiceProviderID)
	if errors.Is(err, types.ErrNoData) {
		return errors.Errorf("service provider not found: id %s", service.ServiceProviderID)
	} else if err != nil {
		return err
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

	now := time.Now()
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
		CreatedAt:        now,
	}

	id, err = uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}
	providerNotification := types.ServiceProviderNotification{
		ID:                id,
		ServiceProviderID: service.ServiceProviderID,
		OfferID:           uuid.NullUUID{UUID: offer.ID, Valid: true},
		Type:              types.ServiceProviderNotificationTypeOfferReceived,
		CreatedAt:         now,
	}

	tx, err := dbUtil.NewSqlxTx(ctx, s.db, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err = s.offerRepo.CreateTx(ctx, tx, offer); err != nil {
		return err
	}

	if err = s.serviceProviderNotificationRepo.CreateTx(ctx, tx, providerNotification); err != nil {
		return err
	}

	key := types.FCMTokenKey(provider.UserID)
	token, err := s.fcmTokenRepo.Find(ctx, key)
	if !errors.Is(err, types.ErrNoData) && err != nil {
		return err
	}

	if token != "" {
		err = s.notificationSvc.SendPush(ctx, types.NotificationSendReq{
			Title:   fmt.Sprintf("%s sent you an offer!", user.Name),
			Message: "Check it now",
			Token:   token,
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

func (s *offerImpl) ConsumerGetAll(ctx context.Context, req types.OfferConsumerGetAllReq) ([]types.OfferConsumerGetAllRes, error) {
	res := []types.OfferConsumerGetAllRes{}

	err := req.Validate()
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

	offers, err := s.offerRepo.FindAllByUserID(ctx, req.AuthUser.ID)
	if err != nil {
		return res, err
	}

	offerIDs := uuid.UUIDs{}
	for _, o := range offers {
		offerIDs = append(offerIDs, o.ID)
	}

	negotiations, err := s.offerNegotiationRepo.FindByOfferIDsAndStatus(ctx, offerIDs, types.OfferNegotiationStatusPending)
	if err != nil {
		return res, err
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
			ID:                    o.ID,
			ServiceCost:           o.ServiceCost,
			ServiceStartDate:      o.ServiceStartDate.Format(time.DateOnly),
			ServiceEndDate:        o.ServiceEndDate.Format(time.DateOnly),
			ServiceStartTime:      o.ServiceStartTime.In(reqTimeZone).Format(time.TimeOnly),
			ServiceEndTime:        o.ServiceEndTime.In(reqTimeZone).Format(time.TimeOnly),
			ServiceTimeTimeZone:   reqTimeZone.String(),
			HasPendingNegotiation: slices.ContainsFunc(negotiations, func(n types.OfferNegotiation) bool { return n.OfferID == o.ID }),
			CreatedAt:             o.CreatedAt,
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

func (s *offerImpl) ConsumerGetByID(ctx context.Context, req types.OfferConsumerGetByIDReq) (types.OfferConsumerGetByIDRes, error) {
	res := types.OfferConsumerGetByIDRes{}

	err := req.Validate()
	if err != nil {
		return res, err
	}

	var timeZone *time.Location
	if req.TimeZone != "" {
		timeZone, err = time.LoadLocation(req.TimeZone)
	} else {
		timeZone, err = time.LoadLocation(types.AppTimeZone)
	}

	if err != nil {
		return res, errors.New(err)
	}

	offer, err := s.offerRepo.FindByIDAndUserID(ctx, req.ID, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "offer not found"})
	} else if err != nil {
		return res, err
	}

	negotiations, err := s.offerNegotiationRepo.FindAllByOfferID(ctx, offer.ID)
	if !errors.Is(err, types.ErrNoData) && err != nil {
		return res, err
	}

	service, err := s.serviceRepo.FindByID(ctx, offer.ServiceID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("service not found: id %s", offer.ServiceID)
	} else if err != nil {
		return res, err
	}

	serviceProvider, err := s.serviceProviderRepo.FindByID(ctx, service.ServiceProviderID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("service provider not found: id %s", service.ServiceProviderID)
	} else if err != nil {
		return res, err
	}

	address, err := s.userAddressRepo.FindByIDAndUserID(ctx, offer.UserAddressID, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("address not found: id %s", offer.UserAddressID)
	} else if err != nil {
		return res, err
	}

	serviceProviderLogoURL, err := s.fileSvc.GetS3PresignedURL(ctx, serviceProvider.LogoImage)
	if err != nil {
		return res, err
	}

	var lat null.Float64
	var lng null.Float64
	if address.Coordinates.Valid {
		latitude, longitude, err := utils.ParseLatLngFromHexStr(address.Coordinates.String)
		if err != nil {
			return res, err
		}

		lat = null.Float64From(latitude)
		lng = null.Float64From(longitude)
	}

	negotiationsRes := []types.OfferConsumerGetByIDResNegotiation{}
	for _, n := range negotiations {
		negotiationsRes = append(negotiationsRes, types.OfferConsumerGetByIDResNegotiation{
			ID:                   n.ID,
			Message:              n.Message,
			RequestedServiceCost: n.RequestedServiceCost,
			Status:               n.Status,
			CreatedAt:            n.CreatedAt,
		})
	}

	res = types.OfferConsumerGetByIDRes{
		ID:                    offer.ID,
		ServiceCost:           offer.ServiceCost,
		Detail:                offer.Detail,
		ServiceStartDate:      offer.ServiceStartDate.Format(time.DateOnly),
		ServiceEndDate:        offer.ServiceEndDate.Format(time.DateOnly),
		ServiceStartTime:      offer.ServiceStartTime.In(timeZone).Format(time.TimeOnly),
		ServiceEndTime:        offer.ServiceEndTime.In(timeZone).Format(time.TimeOnly),
		ServiceTimeTimeZone:   timeZone.String(),
		HasPendingNegotiation: slices.ContainsFunc(negotiations, func(n types.OfferNegotiation) bool { return n.Status == types.OfferNegotiationStatusPending }),
		CreatedAt:             offer.CreatedAt,
		Service: types.OfferConsumerGetByIDResService{
			ID:   service.ServiceProviderID,
			Name: service.Name,
		},
		ServiceProvider: types.OfferConsumerGetByIDResServiceProvider{
			ID:                    serviceProvider.ID,
			Name:                  serviceProvider.Name,
			LogoURL:               serviceProviderLogoURL,
			ReceivedRatingCount:   serviceProvider.ReceivedRatingCount,
			ReceivedRatingAverage: serviceProvider.ReceivedRatingAverage,
		},
		Address: types.OfferConsumerGetByIDResAddress{
			ID:       address.ID,
			Name:     address.Name,
			Province: address.Province,
			City:     address.City,
			Lat:      lat,
			Lng:      lng,
			Address:  address.Address,
		},
		Negotiations: negotiationsRes,
	}

	return res, nil
}

func (s *offerImpl) ProviderAction(ctx context.Context, req types.OfferProviderActionReq) error {
	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.Errorf("service provider not found: user_id %s", req.AuthUser.ID)
	} else if err != nil {
		return err
	}

	offer, err := s.offerRepo.FindByIDAndServiceProviderID(ctx, req.ID, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound, Message: "offer not found"})
	} else if err != nil {
		return err
	}

	if err := req.Validate(offer.ServiceStartDate, offer.ServiceEndDate, offer.ServiceStartTime, offer.ServiceEndTime); err != nil {
		return err
	}

	token, err := s.fcmTokenRepo.Find(ctx, types.FCMTokenKey(offer.UserID))
	if !errors.Is(err, types.ErrNoData) && err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}

	now := time.Now()
	pushNotifReq := types.NotificationSendReq{}
	consumerNotification := types.ConsumerNotification{
		ID:        id,
		UserID:    offer.UserID,
		CreatedAt: now,
	}

	tx, err := dbUtil.NewSqlxTx(ctx, s.db, nil)
	if err != nil {
		errors.New(err)
	}

	defer tx.Rollback()
	switch req.Action {
	case types.OfferProviderActionReqActionAccept:
		offer.Status = types.OfferStatusAccepted
		// TODO: create order

		chatRoomReq := types.ChatChatRoomCreateReq{
			AuthUser:    req.AuthUser,
			SenderID:    req.AuthUser.ID,
			RecipientID: offer.UserID,
			OfferID:     uuid.NullUUID{UUID: offer.ID, Valid: true},
			Tx:          tx,
		}
		_, err = s.chatSvc.CreateChatRoom(ctx, chatRoomReq)
		if err != nil {
			return err
		}

		consumerNotification.Type = types.ConsumerNotificationTypeOfferAccepted
		pushNotifReq = types.NotificationSendReq{
			Title:   fmt.Sprintf("%s accept your offer", provider.Name),
			Message: "confirm your payment now",
			Token:   token,
		}
	case types.OfferProviderActionReqActionReject:
		offer.Status = types.OfferStatusRejected

		consumerNotification.Type = types.ConsumerNotificationTypeOfferRejected
		pushNotifReq = types.NotificationSendReq{
			Title:   fmt.Sprintf("%s reject your offer", provider.Name),
			Message: "your offer has been rejected, you still can sent a new offer :)",
			Token:   token,
		}
	}

	if err := s.consumerNotificationRepo.CreateTx(ctx, tx, consumerNotification); err != nil {
		return err
	}

	if token != "" {
		providerLogoURL, err := s.fileSvc.GetS3PresignedURL(ctx, provider.LogoImage)
		if err != nil {
			return err
		}

		pushNotifReq.ImageURL = providerLogoURL
		go s.notificationSvc.SendPush(ctx, pushNotifReq)
	}

	if err = s.offerRepo.UpdateTx(ctx, tx, offer); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
