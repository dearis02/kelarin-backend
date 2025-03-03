package provider

import (
	"kelarin/internal/service"

	"github.com/google/wire"
)

var ServiceSet = wire.NewSet(
	service.NewUser,
	service.NewAuth,
	service.NewFile,
	service.NewGeocoding,
	service.NewServiceProvider,
	service.NewService,
	service.NewProvince,
	service.NewCity,
	service.NewServiceCategory,
	service.NewConsumerService,
	service.NewUserAddress,
	service.NewOffer,
	service.NewOfferNegotiation,
	service.NewNotification,
)
