package service

import (
	"context"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
)

type ServiceProviderNotification interface {
	GetAll(ctx context.Context, req types.ServiceProviderNotificationGetAllReq) ([]types.ServiceProviderNotificationGetAllRes, error)
}

type serviceProviderNotificationImpl struct {
	serviceProviderRepo             repository.ServiceProvider
	serviceProviderNotificationRepo repository.ServiceProviderNotification
	utilSvc                         Util
}

func NewServiceProviderNotification(serviceProviderRepo repository.ServiceProvider, serviceProviderNotificationRepo repository.ServiceProviderNotification, utilSvc Util) ServiceProviderNotification {
	return &serviceProviderNotificationImpl{
		serviceProviderRepo,
		serviceProviderNotificationRepo,
		utilSvc,
	}
}

func (s *serviceProviderNotificationImpl) GetAll(ctx context.Context, req types.ServiceProviderNotificationGetAllReq) ([]types.ServiceProviderNotificationGetAllRes, error) {
	res := []types.ServiceProviderNotificationGetAllRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(fmt.Sprintf("service provider not found: user_id %s", req.AuthUser.ID))
	} else if err != nil {
		return res, err
	}

	notifications, err := s.serviceProviderNotificationRepo.FindAllByServiceProviderID(ctx, provider.ID)
	if err != nil {
		return res, err
	}

	reqTz, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	for _, n := range notifications {
		details := s.GenerateDetails(n)

		res = append(res, types.ServiceProviderNotificationGetAllRes{
			ID:        n.ID,
			Title:     details.Title,
			Message:   details.Message,
			Read:      n.Read,
			Metadata:  details.Metadata,
			CreatedAt: n.CreatedAt.In(reqTz),
		})
	}

	return res, nil
}

func (s *serviceProviderNotificationImpl) GenerateDetails(notification types.ServiceProviderNotificationWithUser) types.ServiceProviderNotificationGeneratedDetails {
	details := types.ServiceProviderNotificationGeneratedDetails{}

	switch notification.Type {
	case
		types.ServiceProviderNotificationTypeOfferReceived,
		types.ServiceProviderNotificationTypeOfferCanceled:
		details.Metadata = types.ServiceProviderNotificationMetadataOffer{
			OfferID: notification.OfferID.UUID,
		}
	case
		types.ServiceProviderNotificationTypeOfferNegotiationAccepted,
		types.ServiceProviderNotificationTypeOfferNegotiationRejected:
		details.Metadata = types.ServiceProviderNotificationMetadataOfferNegotiation{
			OfferNegotiationID: notification.OfferNegotiationID.UUID,
		}
	case
		types.ServiceProviderNotificationTypeConsumerSettledPayment,
		types.ServiceProviderNotificationTypeOrderFinished:
		details.Metadata = types.ServiceProviderNotificationMetadataOrder{
			OrderID: notification.OrderID.UUID,
		}
	}

	switch notification.Type {
	case types.ServiceProviderNotificationTypeOfferReceived:
		details.Title = fmt.Sprintf("%s sent you an offer", notification.UserName.String)
		details.Message = "You have received an offer. Check it now"
	case types.ServiceProviderNotificationTypeOfferCanceled:
		details.Title = fmt.Sprintf("%s canceled their offer", notification.UserName.String)
		details.Message = "Offer canceled. Contact the consumer for further information"
	case types.ServiceProviderNotificationTypeOfferNegotiationAccepted:
		details.Title = fmt.Sprintf("%s accepted your offer", notification.UserName.String)
		details.Message = "Your offer negotiation has been accepted. Check it now"
	case types.ServiceProviderNotificationTypeOfferNegotiationRejected:
		details.Title = fmt.Sprintf("%s rejected your offer", notification.UserName.String)
		details.Message = "Your offer negotiation has been rejected. Check it now"
	case types.ServiceProviderNotificationTypeOrderFinished:
		details.Title = fmt.Sprintf("%s's order finished", notification.UserName.String)
		details.Message = "Order finished, the service fee automatically added to your credit"
	case types.ServiceProviderNotificationTypeConsumerSettledPayment:
		details.Title = fmt.Sprintf("%s finished their payment for your service fee", notification.UserName.String)
		details.Message = "The service fee is currently on hold!"
	}

	return details
}
