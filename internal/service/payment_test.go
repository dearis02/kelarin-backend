package service_test

import (
	"context"
	"kelarin/internal/config"
	repoMock "kelarin/internal/mocks/repository"
	serviceMock "kelarin/internal/mocks/service"
	"kelarin/internal/service"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"
	"testing"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
)

func TestPaymentService(t *testing.T) {
	db, dbMock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	beginMainDBTx := dbUtil.NewSqlxTx(db)

	ctx := context.Background()

	paymentRepo := repoMock.NewPayment(t)
	paymentMethodRepo := repoMock.NewPaymentMethod(t)
	orderRepo := repoMock.NewOrder(t)
	midtransSvc := serviceMock.NewMidtrans(t)
	notificationSvc := serviceMock.NewNotification(t)
	fcmTokenRepo := repoMock.NewFCMToken(t)
	consumerNotificationRepo := repoMock.NewConsumerNotification(t)
	serviceProviderNotificationRepo := repoMock.NewServiceProviderNotification(t)

	paymentService := service.NewPayment(&config.Config{}, beginMainDBTx, paymentRepo, paymentMethodRepo, orderRepo, midtransSvc, notificationSvc, fcmTokenRepo, consumerNotificationRepo, serviceProviderNotificationRepo)

	amount := decimal.NewFromInt(328000)

	t.Run("Test CalculateAdminFee admin fee unit fixed", func(t *testing.T) {
		total := paymentService.CalculateAdminFee(amount, float32(4500), types.PaymentMethodAdminFeeUnitFixed)
		if !total.Equal(decimal.NewFromInt(4500)) {
			t.Errorf("amount should be 328000, got %d", amount.IntPart())
		}
	})

	t.Run("Test CalculateAdminFee admin fee unit percentage", func(t *testing.T) {
		total := paymentService.CalculateAdminFee(amount, float32(0.8), types.PaymentMethodAdminFeeUnitPercentage)
		if !total.Equal(decimal.NewFromInt(2624)) {
			t.Errorf("total should be 2624, got %s", total.String())
		}
	})

	t.Run("Test Create - payment method enabled", func(t *testing.T) {
		req := types.PaymentCreateReq{
			AuthUser: types.AuthUser{
				ID: uuid.New(),
			},
		}

		adminFee := float32(4500)
		paymentMethodRepo.Mock.On("FindByID", ctx, req.PaymentMethodID).Return(types.PaymentMethod{
			AdminFee:     adminFee,
			AdminFeeUnit: types.PaymentMethodAdminFeeUnitFixed,
			Enabled:      true,
			Code:         string(snap.PaymentTypeBNIVA),
		}, nil)

		serviceFee := decimal.NewFromInt(450000)
		orderRepo.Mock.On("FindByIDAndUserID", ctx, req.OrderID, req.AuthUser.ID).Return(types.OrderWithRelations{
			OfferStatus: types.OfferStatusAccepted,
			Order: types.Order{
				ServiceFee: serviceFee,
			},
		}, nil)

		totalFee := serviceFee.Add(decimal.NewFromFloat32(adminFee)).Add(decimal.NewFromInt(5000))

		mItems := []midtrans.ItemDetails{
			{
				ID:    uuid.Nil.String(),
				Name:  "Service",
				Price: serviceFee.IntPart(),
				Qty:   1,
			},
			{
				Name:  "Admin Fee",
				Price: int64(adminFee),
				Qty:   1,
			},
			{
				Name:  "Platform Fee",
				Price: 5000,
				Qty:   1,
			},
		}

		snapReq := snap.Request{
			TransactionDetails: midtrans.TransactionDetails{
				GrossAmt: totalFee.IntPart(),
			},
			Items:           &mItems,
			EnabledPayments: []snap.SnapPaymentType{snap.SnapPaymentType(snap.PaymentTypeBNIVA)},
			CustomerDetail:  &midtrans.CustomerDetails{},
			Callbacks:       &snap.Callbacks{},
		}

		paymentRedirectURL := "https://midtrans.com"
		midtransSvc.Mock.On("CreateTransaction", ctx, mock.MatchedBy(func(sr *snap.Request) bool {
			return sr.TransactionDetails.GrossAmt == snapReq.TransactionDetails.GrossAmt &&
				sr.EnabledPayments[0] == snapReq.EnabledPayments[0]
		})).Return(&snap.Response{
			RedirectURL: paymentRedirectURL,
		}, nil)

		dbMock.ExpectBegin()
		dbMock.ExpectCommit()

		paymentRepo.Mock.On("CreateTx", ctx, mock.Anything, mock.MatchedBy(func(p types.Payment) bool {
			return p.Amount == serviceFee &&
				p.AdminFee == int32(adminFee) &&
				p.PlatformFee == 5000 &&
				p.PaymentLink == paymentRedirectURL &&
				p.Status == types.PaymentStatusPending
		})).Return(nil)

		orderRepo.Mock.On("UpdateAsPaymentTx", ctx, mock.Anything, mock.MatchedBy(func(o types.Order) bool {
			return o.ServiceFee.Equal(serviceFee)
		})).Return(nil)

		res, err := paymentService.Create(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, paymentRedirectURL, res.PaymentLink)

		orderRepo.AssertExpectations(t)
		paymentRepo.AssertExpectations(t)
		midtransSvc.AssertExpectations(t)

		err = dbMock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
