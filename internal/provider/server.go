package provider

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"
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
	NotificationHandler     handler.Notification
	PaymentHandler          handler.Payment
	OrderHandler            handler.Order
	PaymentMethodHandler    handler.PaymentMethod
	ReportHandler           handler.Report
	ChatHandler             handler.Chat
	AuthMiddleware          middleware.Auth
}

func NewServer(
	userHandler handler.User,
	authHandler *handler.Auth,
	fileHandler handler.File,
	serviceProviderHandler handler.ServiceProvider,
	serviceHandler handler.Service,
	provinceHandler handler.Province,
	cityHandler handler.City,
	serviceCategoryHandler handler.ServiceCategory,
	userAddressHandler handler.UserAddress,
	offerHandler handler.Offer,
	offerNegotiationHandler handler.OfferNegotiation,
	notificationHandler handler.Notification,
	paymentHandler handler.Payment,
	orderHandler handler.Order,
	paymentMethodHandler handler.PaymentMethod,
	reportHandler handler.Report,
	chatHandler handler.Chat,
	authMiddleware middleware.Auth,
) *Server {
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
		notificationHandler,
		paymentHandler,
		orderHandler,
		paymentMethodHandler,
		reportHandler,
		chatHandler,
		authMiddleware,
	}
}
