package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Service interface {
	Create(c *gin.Context)
}

type serviceImpl struct {
	serviceSvc     service.Service
	authMiddleware middleware.Auth
}

func NewService(serviceSvc service.Service, authMiddleware middleware.Auth) Service {
	return &serviceImpl{
		serviceSvc:     serviceSvc,
		authMiddleware: authMiddleware,
	}
}

func (h *serviceImpl) Create(c *gin.Context) {
	var req types.ServiceCreateReq

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.serviceSvc.Create(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		StatusCode: http.StatusCreated,
		Message:    http.StatusText(http.StatusCreated),
	})
}
