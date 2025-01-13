package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceProvider interface {
	Register(c *gin.Context)
}

type serviceProviderImpl struct {
	serviceProviderSvc service.ServiceProvider
	middleware         middleware.Auth
}

func NewServiceProvider(serviceProviderSvc service.ServiceProvider, middleware middleware.Auth) ServiceProvider {
	return &serviceProviderImpl{
		serviceProviderSvc: serviceProviderSvc,
		middleware:         middleware,
	}
}

func (h *serviceProviderImpl) Register(c *gin.Context) {
	var req types.ServiceProviderCreateReq

	if err := h.middleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.serviceProviderSvc.Register(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		StatusCode: http.StatusCreated,
		Message:    http.StatusText(http.StatusCreated),
	})
}
