package service

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"kelarin/internal/config"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"kelarin/internal/utils"
	dbUtil "kelarin/internal/utils/dbutil"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

const PlatformFee = 5000

type Payment interface {
	Create(ctx context.Context, req types.PaymentCreateReq) (types.PaymentCreateRes, error)
	MidtransNotification(ctx context.Context, req types.PaymentMidtransNotificationReq) error
	CalculateAdminFee(amount decimal.Decimal, adminFee float32, adminFeeUnit types.PaymentMethodAdminFeeUnit) decimal.Decimal
}

type paymentImpl struct {
	cfg                             *config.MidtransConfig
	beginMainDBTx                   dbUtil.SqlxTx
	paymentRepo                     repository.Payment
	paymentMethodRepo               repository.PaymentMethod
	orderRepo                       repository.Order
	midtransSvc                     Midtrans
	notificationSvc                 Notification
	fcmTokenRepo                    repository.FCMToken
	consumerNotificationRepo        repository.ConsumerNotification
	serviceProviderNotificationRepo repository.ServiceProviderNotification
}

func NewPayment(cfg *config.Config, beginMainDBTx dbUtil.SqlxTx, paymentRepo repository.Payment, paymentMethodRepo repository.PaymentMethod, orderRepo repository.Order, midtransSvc Midtrans, notificationSvc Notification, fcmTokenRepo repository.FCMToken, consumerNotificationRepo repository.ConsumerNotification, serviceProviderNotificationRepo repository.ServiceProviderNotification) Payment {
	return &paymentImpl{
		cfg:                             &cfg.Midtrans,
		beginMainDBTx:                   beginMainDBTx,
		paymentRepo:                     paymentRepo,
		paymentMethodRepo:               paymentMethodRepo,
		orderRepo:                       orderRepo,
		midtransSvc:                     midtransSvc,
		notificationSvc:                 notificationSvc,
		fcmTokenRepo:                    fcmTokenRepo,
		consumerNotificationRepo:        consumerNotificationRepo,
		serviceProviderNotificationRepo: serviceProviderNotificationRepo,
	}
}

func (s *paymentImpl) Create(ctx context.Context, req types.PaymentCreateReq) (types.PaymentCreateRes, error) {
	res := types.PaymentCreateRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	paymentMethod, err := s.paymentMethodRepo.FindByID(ctx, req.PaymentMethodID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "payment method not found"})
	} else if err != nil {
		return res, err
	}

	if !paymentMethod.Enabled {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "payment method not found"})
	}

	order, err := s.orderRepo.FindByIDAndUserID(ctx, req.OrderID, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "order not found"})
	} else if err != nil {
		return res, err
	}

	if order.OfferStatus != types.OfferStatusAccepted {
		return res, errors.New(types.AppErr{Code: http.StatusForbidden, Message: "offer not accepted yet"})
	}

	if order.PaymentFulfilled {
		return res, errors.New(types.AppErr{Code: http.StatusForbidden, Message: "order already paid"})
	}

	adminFee := s.CalculateAdminFee(order.ServiceFee, paymentMethod.AdminFee, paymentMethod.AdminFeeUnit)
	totalFee := order.ServiceFee.Add(adminFee).Add(decimal.NewFromInt(PlatformFee))

	id, err := uuid.NewV7()
	if err != nil {
		return res, errors.New(err)
	}

	ref := utils.GenerateInvoiceRef(id)

	timeNow := time.Now().Local()

	payment := types.Payment{
		ID:              id,
		Reference:       ref,
		PaymentMethodID: paymentMethod.ID,
		UserID:          order.UserID,
		Amount:          order.ServiceFee,
		AdminFee:        int32(adminFee.IntPart()),
		PlatformFee:     PlatformFee,
		Status:          types.PaymentStatusPending,
		ExpiredAt:       timeNow.Add(time.Hour * 24),
		CreatedAt:       timeNow,
	}

	mItems := []midtrans.ItemDetails{}
	mItems = append(mItems, midtrans.ItemDetails{
		ID:    order.ServiceID.String(),
		Name:  "Service",
		Price: order.ServiceFee.IntPart(),
		Qty:   1,
	})

	mItems = append(mItems, midtrans.ItemDetails{
		Name:  "Admin Fee",
		Price: adminFee.IntPart(),
		Qty:   1,
	})

	mItems = append(mItems, midtrans.ItemDetails{
		Name:  "Platform Fee",
		Price: PlatformFee,
		Qty:   1,
	})

	createPaymentReq := snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  id.String(),
			GrossAmt: totalFee.IntPart(),
		},
		Items:           &mItems,
		EnabledPayments: []snap.SnapPaymentType{snap.SnapPaymentType(paymentMethod.Code)},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: order.UserName,
			Email: order.UserEmail,
		},
		Callbacks: &snap.Callbacks{
			Finish: s.cfg.RedirectURL,
		},
		Expiry: &snap.ExpiryDetails{
			StartTime: timeNow.Format("2006-01-02 15:04:05 Z0700"),
			Unit:      "day",
			Duration:  1,
		},
	}

	createPaymentRes, err := s.midtransSvc.CreateTransaction(ctx, &createPaymentReq)
	if err != nil {
		return res, err
	}

	payment.PaymentLink = createPaymentRes.RedirectURL
	order.PaymentID = uuid.NullUUID{UUID: id, Valid: true}
	order.UpdatedAt = null.TimeFrom(timeNow)

	tx, err := s.beginMainDBTx(ctx, nil)
	if err != nil {
		return res, errors.New(err)
	}

	if err = s.paymentRepo.CreateTx(ctx, tx, payment); err != nil {
		return res, err
	}

	if err = s.orderRepo.UpdateAsPaymentTx(ctx, tx, order.Order); err != nil {
		return res, err
	}

	err = tx.Commit()
	if err != nil {
		return res, err
	}

	res.PaymentLink = createPaymentRes.RedirectURL

	return res, nil
}

func (s *paymentImpl) MidtransNotification(ctx context.Context, req types.PaymentMidtransNotificationReq) error {
	if !s.VerifyMidtransSignatureKey(s.cfg.ServerKey, req) {
		return errors.New(types.AppErr{Code: http.StatusForbidden, Message: "invalid signature key"})
	}

	id, err := uuid.Parse(req.OrderID)
	if err != nil {
		return errors.New(err)
	}

	payment, err := s.paymentRepo.FindByID(ctx, id)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(types.AppErr{Code: http.StatusNotFound, Message: "payment not found"})
	} else if err != nil {
		return err
	}

	// possible multiple payment created
	order, err := s.orderRepo.FindByPaymentID(ctx, payment.ID)
	if errors.Is(err, types.ErrNoData) {
		log.Error().
			Str("payment_id", payment.ID.String()).
			Err(errors.New("order not found")).
			Send()
		return nil
	} else if err != nil {
		return err
	}

	updatedAt := null.TimeFrom(time.Now())
	switch req.TransactionStatus {
	case types.MidtransTransactionStatusPending:
		payment.Status = types.PaymentStatusPending
	case types.MidtransTransactionStatusSettlement:
		payment.Status = types.PaymentStatusPaid
		order.PaymentFulfilled = true
		order.UpdatedAt = updatedAt
		payment.UpdatedAt = updatedAt
	case types.MidtransTransactionStatusExpire:
		payment.Status = types.PaymentStatusExpired
		payment.UpdatedAt = updatedAt
	case types.MidtransTransactionStatusCancel:
		payment.Status = types.PaymentStatusCanceled
		payment.UpdatedAt = updatedAt
	case types.MidtransTransactionStatusFailure, types.MidtransTransactionStatusDeny:
		payment.Status = types.PaymentStatusFailed
		payment.UpdatedAt = updatedAt
	}

	timeNow := time.Now()
	tx, err := s.beginMainDBTx(ctx, nil)
	if err != nil {
		return errors.New(err)
	}

	defer tx.Rollback()

	if err = s.paymentRepo.UpdateStatusTx(ctx, tx, payment); err != nil {
		return err
	}

	if payment.Status == types.PaymentStatusPaid {
		if err = s.orderRepo.UpdateAsPaymentFulfilledTx(ctx, tx, order.Order); err != nil {
			return err
		}

		id, err = uuid.NewV7()
		if err != nil {
			return errors.New(err)
		}
		consumerNotif := types.ConsumerNotification{
			ID:        id,
			UserID:    order.UserID,
			OrderID:   uuid.NullUUID{UUID: order.ID, Valid: true},
			PaymentID: uuid.NullUUID{UUID: payment.ID, Valid: true},
			Type:      types.ConsumerNotificationTypePaymentSuccess,
			CreatedAt: timeNow,
		}

		if err = s.consumerNotificationRepo.CreateTx(ctx, tx, consumerNotif); err != nil {
			return err
		}

		userFCMToken, err := s.fcmTokenRepo.Find(ctx, types.FCMTokenKey(order.UserID))
		if !errors.Is(err, types.ErrNoData) && err != nil {
			return err
		}

		if userFCMToken != "" {
			err = s.notificationSvc.SendPush(ctx, types.NotificationSendReq{
				Title:   "Payment Success",
				Message: "Your order has been paid",
				Token:   userFCMToken,
			})
			if err != nil {
				return err
			}
		}

		id, err = uuid.NewV7()
		if err != nil {
			return errors.New(err)
		}
		providerNotif := types.ServiceProviderNotification{
			ID:                id,
			ServiceProviderID: order.ServiceProviderID,
			OrderID:           uuid.NullUUID{UUID: order.ID, Valid: true},
			Type:              types.ServiceProviderNotificationTypeConsumerSettledPayment,
			CreatedAt:         timeNow,
		}

		if err = s.serviceProviderNotificationRepo.CreateTx(ctx, tx, providerNotif); err != nil {
			return err
		}

		providerFCMToken, err := s.fcmTokenRepo.Find(ctx, types.FCMTokenKey(order.ServiceProviderUserID))
		if !errors.Is(err, types.ErrNoData) && err != nil {
			return err
		}

		if providerFCMToken != "" {
			err = s.notificationSvc.SendPush(ctx, types.NotificationSendReq{
				Title:   fmt.Sprintf("%s has fulfilled the payment", order.UserName),
				Message: "Remember to check the service schedule!",
				Token:   providerFCMToken,
			})
			if err != nil {
				return err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.New(err)
	}

	return nil
}

func (s *paymentImpl) CalculateAdminFee(amount decimal.Decimal, adminFee float32, adminFeeUnit types.PaymentMethodAdminFeeUnit) decimal.Decimal {
	if adminFeeUnit == types.PaymentMethodAdminFeeUnitPercentage {
		return amount.Mul(decimal.NewFromFloat32(adminFee)).Div(decimal.NewFromInt(100)).RoundCeil(0)
	}

	return decimal.NewFromFloat32(adminFee)
}

func (s *paymentImpl) VerifyMidtransSignatureKey(serverKey string, req types.PaymentMidtransNotificationReq) bool {
	payload := req.OrderID + req.StatusCode + req.GrossAmount + serverKey
	hash := sha512.Sum512([]byte(payload))
	expectedSignature := hex.EncodeToString(hash[:])

	return expectedSignature == req.SignatureKey
}
