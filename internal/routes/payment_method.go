package routes

import (
	"kelarin/internal/handler"

	"github.com/gin-gonic/gin"
)

type PaymentMethod struct {
	g                    *gin.Engine
	paymentMethodHandler handler.PaymentMethod
}

func NewPaymentMethod(g *gin.Engine, paymentMethodHandler handler.PaymentMethod) *PaymentMethod {
	return &PaymentMethod{
		g:                    g,
		paymentMethodHandler: paymentMethodHandler,
	}
}

func (r *PaymentMethod) Register() {
	r.g.GET("/common/v1/payment-methods", r.paymentMethodHandler.GetAll)
}
