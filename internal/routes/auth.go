package routes

import (
	"kelarin/internal/handler"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	g           *gin.Engine
	authHandler *handler.Auth
}

func NewAuth(g *gin.Engine, authHandler *handler.Auth) Auth {
	return Auth{
		g:           g,
		authHandler: authHandler,
	}
}

func (r *Auth) Register() {
	r.g.POST("/v1/auth/_login", r.authHandler.Login)
	r.g.POST("/v1/auth/_google_login", r.authHandler.LoginGoogle)
}
