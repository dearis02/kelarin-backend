package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Payment struct {
	g       *gin.Engine
	Payment handler.Payment
}

func NewPayment(g *gin.Engine, paymentHandler handler.Payment) Payment {
	return Payment{
		g:       g,
		Payment: paymentHandler,
	}
}

func (r *Payment) Register(authMw middleware.Auth) {
	r.g.POST("/consumer/v1/payments", authMw.Consumer, r.Payment.Create)

	r.g.POST("/v1/midtrans/notifications", r.Payment.MidtransNotification)
}
