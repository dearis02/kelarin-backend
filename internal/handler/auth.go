package handler

import (
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	authService service.Auth
}

func NewAuth(authService service.Auth) *Auth {
	return &Auth{authService: authService}
}

func (h *Auth) Login(c *gin.Context) {
	var req types.AuthCreateSessionReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.authService.LocalCreateSession(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		Code: http.StatusCreated,
		Data: res,
	})
}

func (h *Auth) ConsumerGoogleLogin(c *gin.Context) {
	var req types.AuthCreateSessionForGoogleReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.authService.ConsumerCreateSession(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		Code: http.StatusCreated,
		Data: res,
	})
}

func (h *Auth) ProviderGoogleLogin(c *gin.Context) {
	var req types.AuthCreateSessionForGoogleReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.authService.ProviderCreateSession(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		Code: http.StatusCreated,
		Data: res,
	})
}
