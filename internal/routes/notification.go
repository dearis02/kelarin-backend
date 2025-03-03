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
}
