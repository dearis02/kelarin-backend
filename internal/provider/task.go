package provider

import (
	"kelarin/internal/repository"
	"kelarin/internal/service"

	"github.com/google/wire"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Cronjob struct {
	db           *sqlx.DB
	redisDB      *redis.Client
	queueClient  *asynq.Client
	OfferService service.Offer
	OrderService service.Order
}

func NewCronjob(
	db *sqlx.DB,
	redisDB *redis.Client,
	queueClient *asynq.Client,
	offerService service.Offer,
	orderService service.Order,
) *Cronjob {
	return &Cronjob{
		db:           db,
		redisDB:      redisDB,
		queueClient:  queueClient,
		OfferService: offerService,
		OrderService: orderService,
	}
}

func (p *Cronjob) Close() error {
	err := p.db.Close()
	if err != nil {
		return err
	}

	err = p.redisDB.Close()
	if err != nil {
		return err
	}

	err = p.queueClient.Close()
	if err != nil {
		return err
	}

	return nil
}

var TaskRepositorySet = wire.NewSet(
	repository.NewUser,
	repository.NewFile,
	repository.NewServiceProvider,
	repository.NewService,
	repository.NewServiceIndex,
	repository.NewUserAddress,
	repository.NewOffer,
	repository.NewOfferNegotiation,
	repository.NewFCMToken,
	repository.NewConsumerNotification,
	repository.NewServiceProviderNotification,
	repository.NewChatRoom,
	repository.NewChatRoomUser,
	repository.NewChatMessage,
	repository.NewOrder,
	repository.NewServiceFeedback,
	repository.NewPayment,
	repository.NewPaymentMethod,
	repository.NewOrderOfferSnapshot,
)

var TaskServiceSet = wire.NewSet(
	service.NewUtil,
	service.NewFile,
	service.NewChat,
	service.NewNotification,
	service.NewOffer,
	service.NewOrder,
)
