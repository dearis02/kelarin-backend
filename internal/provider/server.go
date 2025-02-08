package provider

import (
	"kelarin/internal/handler"
)

type Server struct {
	UserHandler            handler.User
	AuthHandler            *handler.Auth
	FileHandler            handler.File
	ServiceProviderHandler handler.ServiceProvider
	ServiceHandler         handler.Service
	ProvinceHandler        handler.Province
	CityHandler            handler.City
}

func NewServer(userHandler handler.User, authHandler *handler.Auth, fileHandler handler.File, serviceProviderHandler handler.ServiceProvider, serviceHandler handler.Service, provinceHandler handler.Province, cityHandler handler.City) *Server {
	return &Server{
		userHandler,
		authHandler,
		fileHandler,
		serviceProviderHandler,
		serviceHandler,
		provinceHandler,
		cityHandler,
	}
}
