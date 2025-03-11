package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Notification struct {
	g                   *gin.Engine
	notificationHandler handler.Notification
}

func NewNotification(g *gin.Engine, notificationHandler handler.Notification) *Notification {
	return &Notification{
		g:                   g,
		notificationHandler: notificationHandler,
	}
}

func (r *Notification) Register(authMw middleware.Auth) {
	r.g.POST("/v1/notifications/_token", authMw.Authenticated, r.notificationHandler.SaveToken)

	r.g.GET("/consumer/v1/notifications", authMw.Consumer, r.notificationHandler.ConsumerGetAll)

	r.g.GET("/provider/v1/notifications", authMw.ServiceProvider, r.notificationHandler.ProviderGetAll)
}
