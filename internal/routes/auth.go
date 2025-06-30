package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

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

func (r *Auth) Register(authMw middleware.Auth) {
	r.g.POST("/v1/auth/_login", r.authHandler.Login)
	r.g.POST("/consumer/v1/auth/_google_login", r.authHandler.ConsumerGoogleLogin)
	r.g.POST("/provider/v1/auth/_google_login", r.authHandler.ProviderGoogleLogin)
	r.g.POST("/v1/auth/_renew_session", r.authHandler.RenewSession)
	r.g.DELETE("/v1/auth/revoke-session", authMw.Authenticated, r.authHandler.RevokeSession)
}
