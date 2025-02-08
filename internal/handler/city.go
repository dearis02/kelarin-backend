package handler

import (
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
)

type City interface {
	GetByProvinceID(c *gin.Context)
}

type cityImpl struct {
	citySvc service.City
}

func NewCity(citySvc service.City) City {
	return &cityImpl{
		citySvc: citySvc,
	}
}

func (h *cityImpl) GetByProvinceID(c *gin.Context) {
	var req types.CityGetByProvinceIDReq
	var err error

	provinceID := c.Query("province_id")
	req.ProvinceID, err = strconv.ParseInt(provinceID, 10, 64)
	if errors.Is(err, strconv.ErrSyntax) {
		c.Error(errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "province_id is required"}))
		return
	} else if err != nil {
		c.Error(errors.New(types.AppErr{Code: http.StatusBadRequest, Message: err.Error()}))
		return
	}

	res, err := h.citySvc.GetByProvinceID(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}
