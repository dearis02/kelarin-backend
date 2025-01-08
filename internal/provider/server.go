package provider

import (
	"kelarin/internal/handler"
)

type Server struct {
	UserHandler handler.User
	AuthHandler *handler.Auth
}

func NewServer(userHandler handler.User, authHandler *handler.Auth) *Server {
	return &Server{
		userHandler,
		authHandler,
	}
}
