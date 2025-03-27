package service_test

import (
	"context"
	repoMocks "kelarin/internal/mocks/repository"
	svcMocks "kelarin/internal/mocks/service"
	"kelarin/internal/service"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
)

// var appConfig = config.NewApp("../../config/config.yaml")

func TestOfferService(t *testing.T) {
	t.Log("Starting TestOfferService...")

	db, dbMock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	ctx := context.Background()

	offerRepo := repoMocks.NewOffer(t)
	userAddressRepo := repoMocks.NewUserAddress(t)
	serviceRepo := repoMocks.NewService(t)
	serviceProviderRepo := repoMocks.NewServiceProvider(t)
	offerNegotiationRepo := repoMocks.NewOfferNegotiation(t)
	serviceProviderNotificationRepo := repoMocks.NewServiceProviderNotification(t)
	fcmTokenRepo := repoMocks.NewFCMToken(t)
	notificationSvc := svcMocks.NewNotification(t)
	userRepo := repoMocks.NewUser(t)
	consumerNotificationRepo := repoMocks.NewConsumerNotification(t)
	chatSvc := svcMocks.NewChat(t)
	orderRepo := repoMocks.NewOrder(t)
	utilSvc := svcMocks.NewUtil(t)
	fileSvc := svcMocks.NewFile(t)

	timeNow := time.Now()
	serviceStartDate := timeNow.Format(time.DateOnly)
	serviceEndDate := timeNow.Add(time.Hour * 24).Format(time.DateOnly)
	serviceStartTime := timeNow.Format(time.TimeOnly)
	serviceEndTime := timeNow.Add(time.Hour * 4).Format(time.TimeOnly)

	authUserID := uuid.New()
	serviceID := uuid.New()
	addressID := uuid.New()
	serviceProviderID := uuid.New()

	serviceRepo.Mock.On("FindByID", ctx, serviceID).Return(types.Service{
		ID:                serviceID,
		ServiceProviderID: serviceProviderID,
	}, nil)
	offerRepo.Mock.On("IsPendingOfferExists", ctx, authUserID, serviceID).Return(false, nil)
	utilSvc.Mock.On("ParseUserTimeZone", "").Return(time.Local, nil)
	userRepo.Mock.On("FindByID", ctx, authUserID).Return(types.User{ID: authUserID}, nil)
	serviceProviderRepo.Mock.On("FindByID", ctx, serviceProviderID).Return(types.ServiceProvider{
		ID: serviceProviderID,
	}, nil)
	userAddressRepo.Mock.On("FindByIDAndUserID", ctx, addressID, authUserID).Return(types.UserAddress{ID: addressID, UserID: authUserID}, nil)

	localTz := time.FixedZone("GMT+8", 8*60*60)
	startDate, _ := time.Parse(time.DateOnly, serviceStartDate)
	endDate, _ := time.Parse(time.DateOnly, serviceEndDate)
	starTime, _ := time.ParseInLocation(time.TimeOnly, serviceStartTime, localTz)
	endTime, _ := time.ParseInLocation(time.TimeOnly, serviceEndTime, localTz)

	req := types.OfferConsumerCreateReq{
		AuthUser: types.AuthUser{
			ID: authUserID,
		},
		ServiceID:        serviceID,
		AddressID:        addressID,
		Detail:           "foo bar",
		ServiceCost:      100000,
		ServiceStartDate: serviceStartDate,
		ServiceEndDate:   serviceEndDate,
		ServiceStartTime: serviceStartTime,
		ServiceEndTime:   serviceEndTime,
	}

	offerRepo.Mock.On("CreateTx", ctx, mock.Anything, mock.MatchedBy(func(offer types.Offer) bool {
		return offer.UserID == authUserID &&
			offer.ServiceID == req.ServiceID &&
			offer.UserAddressID == req.AddressID &&
			offer.Detail == req.Detail &&
			offer.ServiceCost.Equal(decimal.NewFromFloat(req.ServiceCost)) &&
			offer.ServiceStartDate.Equal(startDate) &&
			offer.ServiceEndDate.Equal(endDate) &&
			offer.ServiceStartTime.Equal(starTime) &&
			offer.ServiceEndTime.Equal(endTime)
	})).Return(nil)

	serviceProviderNotificationRepo.Mock.On("CreateTx", ctx, mock.Anything, mock.MatchedBy(func(n types.ServiceProviderNotification) bool {
		return n.OfferID.Valid && n.Type == types.ServiceProviderNotificationTypeOfferReceived
	})).Return(nil)

	fcmTokenRepo.Mock.On("Find", ctx, types.FCMTokenKey(uuid.UUID{})).Return("", nil)

	dbMock.ExpectBegin()

	beginMainDBTx := dbUtil.NewSqlxTx(db)

	offerService := service.NewOffer(
		beginMainDBTx,
		offerRepo,
		userAddressRepo,
		serviceRepo,
		fileSvc,
		serviceProviderRepo,
		offerNegotiationRepo,
		serviceProviderNotificationRepo,
		fcmTokenRepo,
		notificationSvc,
		userRepo,
		consumerNotificationRepo,
		chatSvc,
		orderRepo,
		utilSvc,
	)

	dbMock.ExpectCommit()
	err = offerService.ConsumerCreate(ctx, req)
	assert.NoError(t, err, "Expected no error when consumer creating an offer")

	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	t.Log("TestOfferService finished")
}
