package provider

import (
	"kelarin/internal/handler"
)

type Server struct {
	UserHandler handler.User
}

func NewServer(userHandler handler.User) *Server {
	return &Server{
		userHandler,
	}
}
