package handler

import (
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Province interface {
	GetAll(c *gin.Context)
}

type provinceImpl struct {
	provinceSv service.Province
}

func NewProvince(provinceSv service.Province) Province {
	return &provinceImpl{
		provinceSv: provinceSv,
	}
}

func (h *provinceImpl) GetAll(c *gin.Context) {
	res, err := h.provinceSv.GetAll(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}
