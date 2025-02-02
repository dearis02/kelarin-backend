package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
)

type Service interface {
	GetAll(c *gin.Context)
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
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

func (h *serviceImpl) GetAll(c *gin.Context) {
	var req types.ServiceGetAllReq

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.serviceSvc.GetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
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

func (h *serviceImpl) GetByID(c *gin.Context) {
	var req types.ServiceGetByIDReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "invalid id"}))
		return
	}

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.serviceSvc.GetByID(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *serviceImpl) Update(c *gin.Context) {
	var req types.ServiceUpdateReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "invalid id"}))
		return
	}

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.serviceSvc.Update(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
	})
}

func (h *serviceImpl) Delete(c *gin.Context) {
	var req types.ServiceDeleteReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "invalid id"}))
		return
	}

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.serviceSvc.Delete(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
	})
}
