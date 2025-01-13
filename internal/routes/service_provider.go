package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type ServiceProvider struct {
	g                      *gin.Engine
	serviceProviderHandler handler.ServiceProvider
}

func NewServiceProvider(g *gin.Engine, serviceProviderHandler handler.ServiceProvider) ServiceProvider {
	return ServiceProvider{
		g:                      g,
		serviceProviderHandler: serviceProviderHandler,
	}
}

func (r *ServiceProvider) Register(m middleware.Auth) {
	r.g.POST("/v1/service-providers", m.ServiceProvider, r.serviceProviderHandler.Register)
}
