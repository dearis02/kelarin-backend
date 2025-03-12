package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Notification interface {
	SaveToken(c *gin.Context)

	ConsumerGetAll(c *gin.Context)

	ProviderGetAll(c *gin.Context)
}

type notificationImpl struct {
	authMw                  middleware.Auth
	notificationSvc         service.Notification
	consumerNotificationSvc service.ConsumerNotification
	providerNotification    service.ServiceProviderNotification
}

func NewNotification(authMw middleware.Auth, notificationSvc service.Notification, consumerNotificationSvc service.ConsumerNotification, providerNotification service.ServiceProviderNotification) Notification {
	return &notificationImpl{
		authMw:                  authMw,
		notificationSvc:         notificationSvc,
		consumerNotificationSvc: consumerNotificationSvc,
		providerNotification:    providerNotification,
	}
}

func (h *notificationImpl) SaveToken(c *gin.Context) {
	var req types.NotificationSaveTokenReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.notificationSvc.SaveToken(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
	})
}

func (h *notificationImpl) ConsumerGetAll(c *gin.Context) {
	var req types.ConsumerNotificationGetAllReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.consumerNotificationSvc.GetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *notificationImpl) ProviderGetAll(c *gin.Context) {
	var req types.ServiceProviderNotificationGetAllReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.providerNotification.GetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}
