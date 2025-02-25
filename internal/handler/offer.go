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
