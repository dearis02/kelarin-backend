package provider

import (
	"kelarin/internal/handler"
)

type Server struct {
	UserHandler handler.User
	AuthHandler *handler.Auth
	FileHandler handler.File
}

func NewServer(userHandler handler.User, authHandler *handler.Auth, fileHandler handler.File) *Server {
	return &Server{
		userHandler,
		authHandler,
		fileHandler,
	}
}
