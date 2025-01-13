package provider

import (
	"kelarin/internal/handler"
)

type Server struct {
	UserHandler            handler.User
	AuthHandler            *handler.Auth
	FileHandler            handler.File
	ServiceProviderHandler handler.ServiceProvider
}

func NewServer(userHandler handler.User, authHandler *handler.Auth, fileHandler handler.File, serviceProviderHandler handler.ServiceProvider) *Server {
	return &Server{
		userHandler,
		authHandler,
		fileHandler,
		serviceProviderHandler,
	}
}
