package handler

import (
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type User interface {
	GetOne(c *gin.Context)
}

type userImpl struct {
	userSvc service.User
}

func NewUser(userSvc service.User) User {
	return &userImpl{
		userSvc: userSvc,
	}
}

func (h *userImpl) GetOne(c *gin.Context) {
	err := h.userSvc.FindOne(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		Code:    http.StatusOK,
		Message: http.StatusText(http.StatusOK),
	})
}
