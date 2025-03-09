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
	ConsumerGetByID(c *gin.Context)
	ConsumerGenerateQRCode(c *gin.Context)

	ProviderGetAll(c *gin.Context)
	ProviderGetByID(c *gin.Context)
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

func (h *orderImpl) ConsumerGetByID(c *gin.Context) {
	var req types.OrderConsumerGetByIDReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(err)
		return
	}

	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.orderSvc.ConsumerGetByID(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *orderImpl) ConsumerGenerateQRCode(c *gin.Context) {
	var req types.OrderConsumerGenerateQRCodeReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(err)
		return
	}

	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.orderSvc.ConsumerGenerateQRCode(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *orderImpl) ProviderGetAll(c *gin.Context) {
	var req types.OrderProviderGetAllReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.orderSvc.ProviderGetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *orderImpl) ProviderGetByID(c *gin.Context) {
	var req types.OrderProviderGetByIDReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(err)
		return
	}

	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.orderSvc.ProviderGetByID(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}
