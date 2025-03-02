package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OfferNegotiation interface {
	ProviderCreate(c *gin.Context)
	ConsumerAction(c *gin.Context)
}

type offerNegotiationImpl struct {
	authMw              middleware.Auth
	offerNegotiationSvc service.OfferNegotiation
}

func NewOfferNegotiation(authMw middleware.Auth, offerNegotiationSvc service.OfferNegotiation) OfferNegotiation {
	return &offerNegotiationImpl{
		authMw:              authMw,
		offerNegotiationSvc: offerNegotiationSvc,
	}
}

func (h *offerNegotiationImpl) ProviderCreate(c *gin.Context) {
	var req types.OfferNegotiationProviderCreateReq
	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.offerNegotiationSvc.ProviderCreate(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.ApiResponse{
		StatusCode: http.StatusCreated,
	})
}

func (h *offerNegotiationImpl) ConsumerAction(c *gin.Context) {
	var req types.OfferNegotiationConsumerActionReq
	if err := req.ID.UnmarshalText([]byte(c.Param("id"))); err != nil {
		c.Error(types.AppErr{Code: http.StatusBadRequest, Message: "invalid id param"})
		return
	}

	if err := h.authMw.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := h.offerNegotiationSvc.ConsumerAction(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
	})
}
