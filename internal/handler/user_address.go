package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserAddress interface {
	Create(c *gin.Context)
	GetAll(c *gin.Context)
	Update(c *gin.Context)
}

type userAddressImpl struct {
	userAddressSvc service.UserAddress
	authMw         middleware.Auth
}

func NewUserAddress(userAddressSvc service.UserAddress, authMw middleware.Auth) UserAddress {
	return &userAddressImpl{
		userAddressSvc: userAddressSvc,
		authMw:         authMw,
	}
}

func (h *userAddressImpl) Create(c *gin.Context) {
	var req types.UserAddressCreateReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.userAddressSvc.Create(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		StatusCode: http.StatusCreated,
	})
}

func (h *userAddressImpl) GetAll(c *gin.Context) {
	var req types.UserAddressGetAllReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.userAddressSvc.GetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *userAddressImpl) Update(c *gin.Context) {
	var req types.UserAddressUpdateReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(types.AppErr{Code: http.StatusBadRequest, Message: "invalid id"})
		return
	}

	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.userAddressSvc.Update(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
	})
}
