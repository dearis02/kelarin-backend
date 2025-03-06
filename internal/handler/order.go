package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Order interface {
	ConsumerGetAll(c *gin.Context)
}

type orderImpl struct {
	orderSvc service.Order
	authMw   middleware.Auth
}

func NewOrder(orderSvc service.Order, authMw middleware.Auth) Order {
	return &orderImpl{
		orderSvc: orderSvc,
		authMw:   authMw,
	}
}

func (h *orderImpl) ConsumerGetAll(c *gin.Context) {
	var req types.OrderConsumerGetAllReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.orderSvc.ConsumerGetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}
