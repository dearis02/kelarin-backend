package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type OfferNegotiation struct {
	g                       *gin.Engine
	offerNegotiationHandler handler.OfferNegotiation
}

func NewOfferNegotiation(g *gin.Engine, offerNegotiationHandler handler.OfferNegotiation) *OfferNegotiation {
	return &OfferNegotiation{
		g:                       g,
		offerNegotiationHandler: offerNegotiationHandler,
	}
}

func (r *OfferNegotiation) Register(authMw middleware.Auth) {
	r.g.POST("/provider/v1/offer-negotiations", authMw.ServiceProvider, r.offerNegotiationHandler.ProviderCreate)

	r.g.PATCH("/consumer/v1/offer-negotiations/:id", authMw.Consumer, r.offerNegotiationHandler.ConsumerAction)
}
