package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Offer interface {
	ConsumerCreate(c *gin.Context)
	ConsumerGetAll(c *gin.Context)
	ConsumerGetByID(c *gin.Context)

	ProviderAction(c *gin.Context)
}

type offerImpl struct {
	offerSvc service.Offer
	authMw   middleware.Auth
}

func NewOffer(offerSvc service.Offer, authMw middleware.Auth) Offer {
	return &offerImpl{
		offerSvc: offerSvc,
		authMw:   authMw,
	}
}

func (h *offerImpl) ConsumerCreate(c *gin.Context) {
	var req types.OfferConsumerCreateReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.offerSvc.ConsumerCreate(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		StatusCode: http.StatusCreated,
	})
}

func (h *offerImpl) ConsumerGetAll(c *gin.Context) {
	var req types.OfferConsumerGetAllReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.offerSvc.ConsumerGetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *offerImpl) ConsumerGetByID(c *gin.Context) {
	var req types.OfferConsumerGetByIDReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(err)
		return
	}

	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.offerSvc.ConsumerGetByID(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *offerImpl) ProviderAction(c *gin.Context) {
	var req types.OfferProviderActionReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(types.AppErr{Code: http.StatusBadRequest, Message: "invalid id param"})
		return
	}

	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	err := h.offerSvc.ProviderAction(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
	})
}
