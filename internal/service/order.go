package service

import (
	"context"
	"fmt"
	"kelarin/internal/config"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v9"
)

type Order interface {
	ConsumerGetAll(ctx context.Context, req types.OrderConsumerGetAllReq) ([]types.OrderConsumerGetAllRes, error)
	ConsumerGetByID(ctx context.Context, req types.OrderConsumerGetByIDReq) (types.OrderConsumerGetByIDRes, error)
	ConsumerGenerateQRCode(ctx context.Context, req types.OrderConsumerGenerateQRCodeReq) (types.OrderConsumerGenerateQRCodeRes, error)

	ProviderGetAll(ctx context.Context, req types.OrderProviderGetAllReq) ([]types.OrderProviderGetAllRes, error)
	ProviderGetByID(ctx context.Context, req types.OrderProviderGetByIDReq) (types.OrderProviderGetByIDRes, error)
	ProviderFinish(ctx context.Context, req types.OrderProviderValidateQRCodeReq) error
}

type orderImpl struct {
	beginMainDBTx                   dbUtil.SqlxTx
	orderRepo                       repository.Order
	fileSvc                         File
	utilSvc                         Util
	offerSvc                        Offer
	paymentRepo                     repository.Payment
	paymentMethodRepo               repository.PaymentMethod
	orderQRCodeSigningKey           string
	serviceProviderRepo             repository.ServiceProvider
	consumerNotificationRepo        repository.ConsumerNotification
	serviceProviderNotificationRepo repository.ServiceProviderNotification
	fcmRepo                         repository.FCMToken
	notificationSvc                 Notification
	serviceRepo                     repository.Service
}

func NewOrder(beginMainDBTx dbUtil.SqlxTx, orderRepo repository.Order, fileSvc File, utilSvc Util, offerSvc Offer, paymentRepo repository.Payment, paymentMethodRepo repository.PaymentMethod, cfg *config.Config, serviceProviderRepo repository.ServiceProvider, consumerNotificationRepo repository.ConsumerNotification,
	serviceProviderNotificationRepo repository.ServiceProviderNotification, fcmRepo repository.FCMToken, notificationSvc Notification, serviceRepo repository.Service) Order {
	return &orderImpl{
		beginMainDBTx:                   beginMainDBTx,
		orderRepo:                       orderRepo,
		fileSvc:                         fileSvc,
		utilSvc:                         utilSvc,
		offerSvc:                        offerSvc,
		paymentRepo:                     paymentRepo,
		paymentMethodRepo:               paymentMethodRepo,
		orderQRCodeSigningKey:           cfg.OrderQRCodeSigningKey,
		serviceProviderRepo:             serviceProviderRepo,
		consumerNotificationRepo:        consumerNotificationRepo,
		serviceProviderNotificationRepo: serviceProviderNotificationRepo,
		fcmRepo:                         fcmRepo,
		notificationSvc:                 notificationSvc,
		serviceRepo:                     serviceRepo,
	}
}

func (s *orderImpl) ConsumerGetAll(ctx context.Context, req types.OrderConsumerGetAllReq) ([]types.OrderConsumerGetAllRes, error) {
	res := []types.OrderConsumerGetAllRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	orders, err := s.orderRepo.FindAllByUserID(ctx, req.AuthUser.ID)
	if err != nil {
		return res, err
	}

	reqTZ, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	for _, order := range orders {
		providerLogoURL, err := s.fileSvc.GetS3PresignedURL(ctx, order.ServiceProviderLogoImage)
		if err != nil {
			return res, err
		}

		var paymentRes *types.OrderConsumerGetAllResPayment
		if order.PaymentID.Valid {
			paymentRes = &types.OrderConsumerGetAllResPayment{
				ID:                order.PaymentID.UUID,
				PaymentMethodName: order.PaymentMethodName.String,
				Amount:            order.PaymentAmount.Decimal,
				AdminFee:          order.PaymentAdminFee.Int32,
				PlatformFee:       order.PaymentPlatformFee.Int32,
				Status:            types.PaymentStatus(order.PaymentStatus.String),
				PaymentLink:       order.PaymentPaymentLink.String,
				CreatedAt:         order.PaymentCreatedAt.Time,
				ExpiredAt:         order.PaymentExpiredAt.Time,
			}
		}

		res = append(res, types.OrderConsumerGetAllRes{
			ID:               order.ID,
			OfferID:          order.OfferID,
			PaymentFulfilled: order.PaymentFulfilled,
			ServiceFee:       order.ServiceFee,
			ServiceDate:      order.ServiceDate.Format(time.DateOnly),
			ServiceTime:      s.utilSvc.NormalizeTimeOnlyTz(order.ServiceTime).In(reqTZ).Format(time.TimeOnly),
			Status:           order.Status,
			CreatedAt:        order.CreatedAt,
			Service: types.OrderConsumerGetAllResService{
				ID:   order.ServiceID,
				Name: order.ServiceName,
			},
			ServiceProvider: types.OrderConsumerGetAllResServiceProvider{
				ID:      order.ServiceProviderID,
				Name:    order.ServiceProviderName,
				LogoURL: providerLogoURL,
			},
			Payment: paymentRes,
		})
	}

	return res, nil
}

func (s *orderImpl) ConsumerGetByID(ctx context.Context, req types.OrderConsumerGetByIDReq) (types.OrderConsumerGetByIDRes, error) {
	res := types.OrderConsumerGetByIDRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	order, err := s.orderRepo.FindByIDAndUserID(ctx, req.ID, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "order not found"})
	} else if err != nil {
		return res, err
	}

	offer, err := s.offerSvc.ConsumerGetByID(ctx, types.OfferConsumerGetByIDReq{ID: order.OfferID, AuthUser: req.AuthUser, TimeZone: req.TimeZone})
	if err != nil {
		return res, err
	}

	var paymentRes *types.OrderConsumerGetByIDResPayment
	if order.PaymentID.Valid {
		payment, err := s.paymentRepo.FindByID(ctx, order.PaymentID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.Errorf("payment not found: id %s", order.PaymentID.UUID)
		} else if err != nil {
			return res, err
		}

		paymentMethod, err := s.paymentMethodRepo.FindByID(ctx, payment.PaymentMethodID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.Errorf("payment method not found: id %s", payment.PaymentMethodID)
		} else if err != nil {
			return res, err
		}

		paymentRes = &types.OrderConsumerGetByIDResPayment{
			ID:                payment.ID,
			Reference:         payment.Reference,
			PaymentMethodName: paymentMethod.Name,
			PaymentMethodLogo: paymentMethod.Logo,
			Amount:            payment.Amount,
			AdminFee:          payment.AdminFee,
			PlatformFee:       payment.PlatformFee,
			Status:            payment.Status,
			PaymentLink:       payment.PaymentLink,
			ExpiredAt:         payment.ExpiredAt,
			CreatedAt:         payment.CreatedAt,
			UpdatedAt:         payment.UpdatedAt,
		}
	}

	reqTz, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	res = types.OrderConsumerGetByIDRes{
		ID:               order.ID,
		OfferID:          order.OfferID,
		PaymentFulfilled: order.PaymentFulfilled,
		ServiceFee:       order.ServiceFee,
		ServiceDate:      order.ServiceDate.Format(time.DateOnly),
		ServiceTime:      s.utilSvc.NormalizeTimeOnlyTz(order.ServiceTime).In(reqTz).Format(time.TimeOnly),
		Status:           order.Status,
		CreatedAt:        order.CreatedAt,
		Offer:            offer,
		Payment:          paymentRes,
	}

	return res, nil
}

func (s *orderImpl) ConsumerGenerateQRCode(ctx context.Context, req types.OrderConsumerGenerateQRCodeReq) (types.OrderConsumerGenerateQRCodeRes, error) {
	res := types.OrderConsumerGenerateQRCodeRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	order, err := s.orderRepo.FindByIDAndUserID(ctx, req.ID, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "order not found"})
	} else if err != nil {
		return res, err
	}

	if !order.PaymentFulfilled {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "order not found"})
	}

	if order.Status != types.OrderStatusOngoing {
		return res, errors.New(types.AppErr{Code: http.StatusForbidden, Message: "order not ongoing"})
	}

	payment, err := s.paymentRepo.FindByID(ctx, order.PaymentID.UUID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("payment not found: id %s", order.PaymentID.UUID)
	} else if err != nil {
		return res, err
	}

	duration := time.Minute * 1
	qrCodeContent, err := s.GenerateQRCodeContent(types.OrderConsumerGenerateQRCodePayload{
		OrderID:     order.ID,
		Amount:      payment.Amount,
		AdminFee:    payment.AdminFee,
		PlatformFee: payment.PlatformFee,
	}, duration)
	if err != nil {
		return res, err
	}

	res = types.OrderConsumerGenerateQRCodeRes{
		QRCodeContent:         qrCodeContent,
		ValidDurationInSecond: duration.Seconds(),
	}

	return res, nil
}

func (s *orderImpl) GenerateQRCodeContent(payload types.OrderConsumerGenerateQRCodePayload, expiration time.Duration) (string, error) {
	res := ""

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, types.OrderConsumerGenerateQRCodePayload{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		},
		OrderID:     payload.OrderID,
		Amount:      payload.Amount,
		AdminFee:    payload.AdminFee,
		PlatformFee: payload.PlatformFee,
	}).SignedString([]byte(s.orderQRCodeSigningKey))
	if err != nil {
		return res, errors.New(err)
	}

	return token, nil
}

func (s *orderImpl) ProviderGetAll(ctx context.Context, req types.OrderProviderGetAllReq) ([]types.OrderProviderGetAllRes, error) {
	res := []types.OrderProviderGetAllRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("service provider not found: user_id %s", req.AuthUser.ID)
	} else if err != nil {
		return res, err
	}

	orders, err := s.orderRepo.FindAllByServiceProviderID(ctx, provider.ID)
	if err != nil {
		return res, err
	}

	paymentIDs := uuid.UUIDs{}
	for _, order := range orders {
		if order.PaymentID.Valid {
			paymentIDs = append(paymentIDs, order.PaymentID.UUID)
		}
	}

	payments, err := s.paymentRepo.FindByIDs(ctx, paymentIDs)
	if err != nil {
		return res, err
	}

	reqTz, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	for _, order := range orders {
		var paymentRes *types.OrderProviderGetAllResPayment

		for _, payment := range payments {
			if order.PaymentID.Valid && order.PaymentID.UUID == payment.ID {
				paymentRes = &types.OrderProviderGetAllResPayment{
					ID:                payment.ID,
					PaymentMethodName: payment.PaymentMethodName,
					Amount:            payment.Amount,
					AdminFee:          payment.AdminFee,
					PlatformFee:       payment.PlatformFee,
					Status:            payment.Status,
				}
			}
		}

		res = append(res, types.OrderProviderGetAllRes{
			ID:               order.ID,
			OfferID:          order.OfferID,
			PaymentFulfilled: order.PaymentFulfilled,
			ServiceFee:       order.ServiceFee,
			ServiceDate:      order.ServiceDate.Format(time.DateOnly),
			ServiceTime:      s.utilSvc.NormalizeTimeOnlyTz(order.ServiceTime).In(reqTz).Format(time.TimeOnly),
			Status:           order.Status,
			CreatedAt:        order.CreatedAt,
			Payment:          paymentRes,
		})
	}

	return res, nil
}

func (s *orderImpl) ProviderGetByID(ctx context.Context, req types.OrderProviderGetByIDReq) (types.OrderProviderGetByIDRes, error) {
	res := types.OrderProviderGetByIDRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("service provider not found: user_id %s", req.AuthUser.ID)
	} else if err != nil {
		return res, err
	}

	order, err := s.orderRepo.FindByIDAndServiceProviderID(ctx, req.ID, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "order not found"})
	} else if err != nil {
		return res, err
	}

	offer, err := s.offerSvc.ProviderGetByID(ctx, types.OfferProviderGetByIDReq{ID: order.OfferID, AuthUser: req.AuthUser})
	if err != nil {
		return res, err
	}

	service, err := s.serviceRepo.FindByID(ctx, offer.ServiceID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("service not found: id %s", offer.ServiceID)
	} else if err != nil {
		return res, err
	}

	res = types.OrderProviderGetByIDRes{
		ID:               order.ID,
		OfferID:          order.OfferID,
		PaymentFulfilled: order.PaymentFulfilled,
		ServiceName:      service.Name,
		ServiceFee:       order.ServiceFee,
		ServiceDate:      order.ServiceDate.Format(time.DateOnly),
		ServiceTime:      s.utilSvc.NormalizeTimeOnlyTz(order.ServiceTime).Format(time.TimeOnly),
		Status:           order.Status,
		CreatedAt:        order.CreatedAt,
		User: types.OrderProviderGetByIDResUser{
			ID:   offer.User.ID,
			Name: offer.User.Name,
		},
		Offer: types.OrderProviderGetByIDResOffer{
			ID:     offer.ID,
			Detail: offer.Detail,
		},
		Address: types.OrderProviderGetByIDResAddress{
			ID:       offer.User.Address.ID,
			Province: offer.User.Address.Province,
			City:     offer.User.Address.City,
			Lat:      offer.User.Address.Lat,
			Lng:      offer.User.Address.Lng,
			Address:  offer.User.Address.Address,
		},
	}

	if order.PaymentID.Valid {
		payment, err := s.paymentRepo.FindByID(ctx, order.PaymentID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.Errorf("payment not found: id %s", order.PaymentID.UUID)
		}

		paymentMethod, err := s.paymentMethodRepo.FindByID(ctx, payment.PaymentMethodID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.Errorf("payment method not found: id %s", payment.PaymentMethodID)
		} else if err != nil {
			return res, err
		}

		res.Payment = &types.OrderProviderGetAllResPayment{
			ID:                payment.ID,
			PaymentMethodName: paymentMethod.Name,
			Amount:            payment.Amount,
			AdminFee:          payment.AdminFee,
			PlatformFee:       payment.PlatformFee,
			Status:            payment.Status,
		}
	}

	return res, nil
}

func (s *orderImpl) ProviderFinish(ctx context.Context, req types.OrderProviderValidateQRCodeReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	claims := &types.OrderConsumerGenerateQRCodePayload{}
	token, err := jwt.ParseWithClaims(req.QRCodeContent, claims, func(token *jwt.Token) (any, error) {
		return []byte(s.orderQRCodeSigningKey), nil
	})
	if errors.Is(err, jwt.ErrTokenExpired) {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "qr-code expired"})
	} else if err != nil {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "invalid qr-code"})
	}

	if !token.Valid {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "invalid qr-code"})
	}

	claims, ok := token.Claims.(*types.OrderConsumerGenerateQRCodePayload)
	if !ok {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "invalid qr-code payload"})
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.Errorf("service provider not found: user_id %s", req.AuthUser.ID)
	} else if err != nil {
		return err
	}

	order, err := s.orderRepo.FindByIDAndServiceProviderID(ctx, claims.OrderID, provider.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound, Message: "order not found"})
	} else if err != nil {
		return err
	}

	payment, err := s.paymentRepo.FindByID(ctx, order.PaymentID.UUID)
	if errors.Is(err, types.ErrNoData) {
		return errors.Errorf("payment not found: id %s", order.PaymentID.UUID)
	} else if err != nil {
		return err
	}

	if order.Status != types.OrderStatusOngoing {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "invalid order"})
	}

	if !order.PaymentFulfilled {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "payment not fulfilled"})
	}

	if claims.AdminFee != payment.AdminFee || !claims.Amount.Equal(payment.Amount) || claims.PlatformFee != payment.PlatformFee {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "invalid qr-code"})
	}

	timeNow := time.Now()

	id, err := uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}
	consumerNotif := types.ConsumerNotification{
		ID:        id,
		UserID:    order.UserID,
		OrderID:   uuid.NullUUID{UUID: order.ID, Valid: true},
		Type:      types.ConsumerNotificationTypeOrderFinished,
		CreatedAt: timeNow,
	}

	id, err = uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}
	providerNotif := types.ServiceProviderNotification{
		ID:                id,
		ServiceProviderID: provider.ID,
		OrderID:           uuid.NullUUID{UUID: order.ID, Valid: true},
		Type:              types.ServiceProviderNotificationTypeOrderFinished,
		CreatedAt:         timeNow,
	}

	tx, err := s.beginMainDBTx(ctx, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	order.Status = types.OrderStatusFinished
	order.UpdatedAt = null.TimeFrom(timeNow)
	if err = s.orderRepo.UpdateStatusTx(ctx, tx, order); err != nil {
		return err
	}

	provider.Credit = provider.Credit.Add(order.ServiceFee)
	if err = s.serviceProviderRepo.UpdateCreditTx(ctx, provider); err != nil {
		return err
	}

	if err = s.consumerNotificationRepo.CreateTx(ctx, tx, consumerNotif); err != nil {
		return err
	}

	if err = s.serviceProviderNotificationRepo.CreateTx(ctx, tx, providerNotif); err != nil {
		return err
	}

	consumerFCMToken, err := s.fcmRepo.Find(ctx, types.FCMTokenKey(order.UserID))
	if !errors.Is(err, types.ErrNoData) && err != nil {
		return err
	}

	providerFCMToken, err := s.fcmRepo.Find(ctx, types.FCMTokenKey(provider.UserID))
	if !errors.Is(err, types.ErrNoData) && err != nil {
		return err
	}

	if consumerFCMToken != "" {
		err = s.notificationSvc.SendPush(ctx, types.NotificationSendReq{
			Title:   fmt.Sprintf("%s order finished", provider.Name),
			Message: "rate provider now",
			Token:   consumerFCMToken,
		})
		if err != nil {
			return err
		}
	}

	if providerFCMToken != "" {
		err = s.notificationSvc.SendPush(ctx, types.NotificationSendReq{
			Title:   "Order finished",
			Message: "the service fee has been added to your credit",
			Token:   providerFCMToken,
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
