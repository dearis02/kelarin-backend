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
