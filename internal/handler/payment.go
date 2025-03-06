package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Payment interface {
	Create(c *gin.Context)
	MidtransNotification(c *gin.Context)
}

type paymentImpl struct {
	paymentSvc service.Payment
	authMw     middleware.Auth
}

func NewPayment(paymentSvc service.Payment, authMw middleware.Auth) Payment {
	return &paymentImpl{
		paymentSvc: paymentSvc,
		authMw:     authMw,
	}
}

func (h *paymentImpl) Create(c *gin.Context) {
	var req types.PaymentCreateReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.paymentSvc.Create(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		StatusCode: http.StatusCreated,
		Data:       res,
	})
}

func (h *paymentImpl) MidtransNotification(c *gin.Context) {
	var req types.PaymentMidtransNotificationReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if err := h.paymentSvc.MidtransNotification(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
	})
}
