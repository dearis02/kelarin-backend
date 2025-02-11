package handler

import (
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceCategory interface {
	GetAll(c *gin.Context)
}

type serviceCategoryImpl struct {
	serviceCategorySvc service.ServiceCategory
}

func NewServiceCategory(serviceCategorySvc service.ServiceCategory) ServiceCategory {
	return &serviceCategoryImpl{
		serviceCategorySvc: serviceCategorySvc,
	}
}

func (h *serviceCategoryImpl) GetAll(c *gin.Context) {
	res, err := h.serviceCategorySvc.GetAll(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}
