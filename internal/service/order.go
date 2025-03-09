package service

import (
	"context"
	"kelarin/internal/config"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/golang-jwt/jwt/v5"
)

type Order interface {
	ConsumerGetAll(ctx context.Context, req types.OrderConsumerGetAllReq) ([]types.OrderConsumerGetAllRes, error)
	ConsumerGetByID(ctx context.Context, req types.OrderConsumerGetByIDReq) (types.OrderConsumerGetByIDRes, error)
	ConsumerGenerateQRCode(ctx context.Context, req types.OrderConsumerGenerateQRCodeReq) (types.OrderConsumerGenerateQRCodeRes, error)
}

type orderImpl struct {
	orderRepo             repository.Order
	fileSvc               File
	utilSvc               Util
	offerSvc              Offer
	paymentRepo           repository.Payment
	paymentMethodRepo     repository.PaymentMethod
	orderQRCodeSigningKey string
}

func NewOrder(orderRepo repository.Order, fileSvc File, utilSvc Util, offerSvc Offer, paymentRepo repository.Payment, paymentMethodRepo repository.PaymentMethod, cfg *config.Config) Order {
	return &orderImpl{
		orderRepo:             orderRepo,
		fileSvc:               fileSvc,
		utilSvc:               utilSvc,
		offerSvc:              offerSvc,
		paymentRepo:           paymentRepo,
		paymentMethodRepo:     paymentMethodRepo,
		orderQRCodeSigningKey: cfg.OrderQRCodeSigningKey,
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
			PaymentMethodName: paymentMethod.Name,
			Amount:            payment.Amount,
			AdminFee:          payment.AdminFee,
			PlatformFee:       payment.PlatformFee,
			Status:            payment.Status,
			PaymentLink:       payment.PaymentLink,
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
