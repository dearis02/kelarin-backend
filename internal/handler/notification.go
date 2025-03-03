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
}

type notificationImpl struct {
	notificationSvc service.Notification
	authMw          middleware.Auth
}

func NewNotification(notificationSvc service.Notification, authMw middleware.Auth) Notification {
	return &notificationImpl{
		notificationSvc: notificationSvc,
		authMw:          authMw,
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
