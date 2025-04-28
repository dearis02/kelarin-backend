package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Service struct {
	g              *gin.Engine
	serviceHandler handler.Service
}

func NewService(g *gin.Engine, serviceHandler handler.Service) Service {
	return Service{
		g:              g,
		serviceHandler: serviceHandler,
	}
}

func (r *Service) Register(m middleware.Auth) {
	r.g.GET("/provider/v1/services", m.ServiceProvider, r.serviceHandler.GetAll)
	r.g.POST("/provider/v1/services", m.ServiceProvider, r.serviceHandler.Create)
	r.g.GET("/provider/v1/services/:id", m.ServiceProvider, r.serviceHandler.GetByID)
	r.g.PUT("/provider/v1/services/:id", m.ServiceProvider, r.serviceHandler.Update)
	r.g.DELETE("/provider/v1/services/:id", m.ServiceProvider, r.serviceHandler.Delete)
	r.g.POST("/provider/v1/services/:id/_images", m.ServiceProvider, r.serviceHandler.AddImages)
	r.g.DELETE("/provider/v1/services/:id/_images", m.ServiceProvider, r.serviceHandler.RemoveImages)

	r.g.GET("/v1/services", r.serviceHandler.ConsumerGetAll)
	r.g.GET("/v1/services/:id", r.serviceHandler.ConsumerGetByID)

	r.g.POST("/consumer/v1/service-feedbacks", m.Consumer, r.serviceHandler.ConsumerCreateServiceFeedback)
}
