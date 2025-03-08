package handler

import (
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentMethod interface {
	GetAll(c *gin.Context)
}

type paymentMethodImpl struct {
	paymentMethodSvc service.PaymentMethod
}

func NewPaymentMethod(paymentMethodSvc service.PaymentMethod) PaymentMethod {
	return &paymentMethodImpl{paymentMethodSvc: paymentMethodSvc}
}

func (p *paymentMethodImpl) GetAll(c *gin.Context) {
	res, err := p.paymentMethodSvc.GetAll(c)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}
