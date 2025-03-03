package provider

import (
	"kelarin/internal/repository"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	repository.NewUser,
	repository.NewSession,
	repository.NewPendingRegistration,
	repository.NewFile,
	repository.NewProvince,
	repository.NewCity,
	repository.NewServiceProviderArea,
	repository.NewServiceProvider,
	repository.NewServiceCategory,
	repository.NewServiceServiceCategory,
	repository.NewService,
	repository.NewServiceIndex,
	repository.NewUserAddress,
	repository.NewOffer,
	repository.NewOfferNegotiation,
	repository.NewFCMToken,
	repository.NewConsumerNotification,
)
