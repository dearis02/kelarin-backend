package provider

import (
	"kelarin/internal/handler"
)

type Server struct {
	UserHandler             handler.User
	AuthHandler             *handler.Auth
	FileHandler             handler.File
	ServiceProviderHandler  handler.ServiceProvider
	ServiceHandler          handler.Service
	ProvinceHandler         handler.Province
	CityHandler             handler.City
	ServiceCategoryHandler  handler.ServiceCategory
	UserAddressHandler      handler.UserAddress
	OfferHandler            handler.Offer
	OfferNegotiationHandler handler.OfferNegotiation
}

func NewServer(userHandler handler.User, authHandler *handler.Auth, fileHandler handler.File, serviceProviderHandler handler.ServiceProvider, serviceHandler handler.Service, provinceHandler handler.Province, cityHandler handler.City, serviceCategoryHandler handler.ServiceCategory, userAddressHandler handler.UserAddress, offerHandler handler.Offer, offerNegotiationHandler handler.OfferNegotiation) *Server {
	return &Server{
		userHandler,
		authHandler,
		fileHandler,
		serviceProviderHandler,
		serviceHandler,
		provinceHandler,
		cityHandler,
		serviceCategoryHandler,
		userAddressHandler,
		offerHandler,
		offerNegotiationHandler,
	}
}
