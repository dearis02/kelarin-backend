package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Order struct {
	g            *gin.Engine
	orderHandler handler.Order
}

func NewOrder(g *gin.Engine, orderHandler handler.Order) *Order {
	return &Order{
		g:            g,
		orderHandler: orderHandler,
	}
}

func (r *Order) Register(authMw middleware.Auth) {
	r.g.GET("/consumer/v1/orders", authMw.Consumer, r.orderHandler.ConsumerGetAll)
	r.g.GET("/consumer/v1/orders/:id", authMw.Consumer, r.orderHandler.ConsumerGetByID)
	r.g.POST("/consumer/v1/orders/:id", authMw.Consumer, r.orderHandler.ConsumerGenerateQRCode)

	r.g.GET("/provider/v1/orders", authMw.ServiceProvider, r.orderHandler.ProviderGetAll)
	r.g.GET("/provider/v1/orders/:id", authMw.ServiceProvider, r.orderHandler.ProviderGetByID)
}
