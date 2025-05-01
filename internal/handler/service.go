package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type Service interface {
	GetAll(c *gin.Context)
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	AddImages(c *gin.Context)
	RemoveImages(c *gin.Context)

	ConsumerGetAll(c *gin.Context)
	ConsumerGetByID(c *gin.Context)
	ConsumerCreateServiceFeedback(c *gin.Context)
}

type serviceImpl struct {
	serviceSvc     service.Service
	consumerSvc    service.ConsumerService
	authMiddleware middleware.Auth
}

func NewService(serviceSvc service.Service, consumerSvc service.ConsumerService, authMiddleware middleware.Auth) Service {
	return &serviceImpl{
		serviceSvc:     serviceSvc,
		consumerSvc:    consumerSvc,
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

func (h *serviceImpl) AddImages(c *gin.Context) {
	var req types.ServiceImageActionReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "invalid id"}))
		return
	}

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.serviceSvc.AddImages(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
	})
}

func (h *serviceImpl) RemoveImages(c *gin.Context) {
	var req types.ServiceImageActionReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "invalid id"}))
		return
	}

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.serviceSvc.RemoveImages(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
	})
}

func (h *serviceImpl) ConsumerGetAll(c *gin.Context) {
	var req types.ConsumerServiceGetAllReq
	var err error

	req.After = c.Query("after")
	req.Keyword = c.Query("keyword")
	req.Province = c.Query("province")
	req.City = c.Query("city")
	req.Size = c.Query("size")
	req.Page = c.Query("page")
	categories := c.QueryArray("categories")
	categories = lo.Filter(categories, func(category string, _ int) bool {
		return category != ""
	})

	req.Categories = categories

	res, paginationRes, err := h.consumerSvc.GetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
		Pagination: &paginationRes,
	})
}

func (h *serviceImpl) ConsumerGetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(types.AppErr{Code: http.StatusBadRequest, Message: "invalid id param"})
		return
	}

	res, err := h.consumerSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *serviceImpl) ConsumerCreateServiceFeedback(c *gin.Context) {
	var req types.ConsumerServiceFeedbackCreateReq
	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.consumerSvc.CreateFeedback(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		StatusCode: http.StatusCreated,
	})
}
