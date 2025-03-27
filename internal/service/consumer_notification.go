package service

import (
	"context"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"kelarin/internal/utils"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/text/currency"
)

type ConsumerNotification interface {
	Create(ctx context.Context, tx dbUtil.Tx, req types.ConsumerNotification) error
	GetAll(ctx context.Context, req types.ConsumerNotificationGetAllReq) ([]types.ConsumerNotificationGetAllRes, error)
}

type consumerNotificationImpl struct {
	beginMainDBTx            dbUtil.SqlxTx
	userRepo                 repository.User
	consumerNotificationRepo repository.ConsumerNotification
	utilSvc                  Util
	fileSvc                  File
}

func NewConsumerNotification(beginMainDBTx dbUtil.SqlxTx, userRepo repository.User, consumerNotificationRepo repository.ConsumerNotification, utilSvc Util, fileSvc File) ConsumerNotification {
	return &consumerNotificationImpl{
		beginMainDBTx:            beginMainDBTx,
		userRepo:                 userRepo,
		consumerNotificationRepo: consumerNotificationRepo,
		utilSvc:                  utilSvc,
		fileSvc:                  fileSvc,
	}
}

func (s *consumerNotificationImpl) Create(ctx context.Context, _tx dbUtil.Tx, req types.ConsumerNotification) error {
	var err error

	tx := _tx
	if _tx == nil {
		tx, err = s.beginMainDBTx(ctx, nil)
		if err != nil {
			return errors.New(err)
		}

		defer tx.Rollback()
	}

	if err = s.consumerNotificationRepo.CreateTx(ctx, tx, req); err != nil {
		return err
	}

	if _tx == nil {
		err = tx.Commit()
		if err != nil {
			return errors.New(err)
		}
	}

	return nil
}

func (s *consumerNotificationImpl) GetAll(ctx context.Context, req types.ConsumerNotificationGetAllReq) ([]types.ConsumerNotificationGetAllRes, error) {
	res := []types.ConsumerNotificationGetAllRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	notification, err := s.consumerNotificationRepo.FindAllByUserID(ctx, req.AuthUser.ID)
	if err != nil {
		return res, err
	}

	reqTz, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	for _, n := range notification {
		details := s.GenerateDetails(n)

		providerLogoURL := ""
		if n.ServiceProviderLogoImage.Valid {
			providerLogoURL, err = s.fileSvc.GetS3PresignedURL(ctx, n.ServiceProviderLogoImage.String)
			if err != nil {
				return res, err
			}
		}

		res = append(res, types.ConsumerNotificationGetAllRes{
			ID:                     n.ID,
			Title:                  details.Title,
			Message:                details.Message,
			ServiceProviderLogoURL: providerLogoURL,
			Read:                   n.Read,
			CreatedAt:              n.CreatedAt.In(reqTz),
			Metadata:               details.Metadata,
		})
	}

	return res, nil
}

func (s *consumerNotificationImpl) GenerateDetails(notification types.ConsumerNotificationWithServiceProviderAndPayment) types.ConsumerNotificationGeneratedDetails {
	details := types.ConsumerNotificationGeneratedDetails{}

	switch notification.Type {
	case types.ConsumerNotificationTypeOfferNegotiationReceived:
		details.Metadata = types.ConsumerNotificationMetadataOfferNegotiation{
			OfferNegotiationID: notification.OfferNegotiationID.UUID,
		}
	case types.ConsumerNotificationTypeOfferAccepted,
		types.ConsumerNotificationTypeOfferRejected:
		details.Metadata = types.ConsumerNotificationMetadataOffer{
			OfferID: notification.OfferID.UUID,
		}
	case types.ConsumerNotificationTypePaymentSuccess,
		types.ConsumerNotificationTypePaymentExpired:
		details.Metadata = types.ConsumerNotificationMetadataPayment{
			PaymentID: notification.PaymentID.UUID,
		}
	case types.ConsumerNotificationTypeOrderFinished:
		details.Metadata = types.ConsumerNotificationMetadataOrder{
			OrderID: notification.OrderID.UUID,
		}
	}

	switch notification.Type {
	case types.ConsumerNotificationTypeOfferNegotiationReceived:
		details.Title = fmt.Sprintf("Offer negotiation received from %s", notification.ServiceProviderName.String)
		details.Message = "You have received an offer negotiation. Please check your offer"
	case types.ConsumerNotificationTypeOfferAccepted:
		details.Title = fmt.Sprintf("%s accepted your offer", notification.ServiceProviderName.String)
		details.Message = "Your offer has been accepted"
	case types.ConsumerNotificationTypeOfferRejected:
		details.Title = fmt.Sprintf("%s rejected your offer", notification.ServiceProviderName.String)
		details.Message = "Your offer has been rejected"
	case types.ConsumerNotificationTypePaymentSuccess:
		amount := notification.PaymentAmount.Decimal.Add(decimal.NewFromInt32(notification.PaymentAdminFee.Int32).Add(decimal.NewFromInt32(notification.PaymentPlatformFee.Int32)))

		details.Title = "Payment success"
		details.Message = fmt.Sprintf("You has paid %s with %s", utils.FormatRupiah(currency.IDR.Amount(amount.InexactFloat64())), notification.PaymentMethodName.String)
	case types.ConsumerNotificationTypePaymentExpired:
		details.Title = "Your payment is expired"
		details.Message = fmt.Sprintf("Your payment with order id %s is expire, create a new one!", notification.OrderID.UUID.String())
	case types.ConsumerNotificationTypeOrderFinished:
		details.Title = fmt.Sprintf("%s's order finished", notification.ServiceProviderName.String)
		details.Message = "Your order has been finished. Rate service provider now!"
	}

	return details
}
