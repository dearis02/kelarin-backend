package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Offer struct {
	g            *gin.Engine
	offerHandler handler.Offer
}

func NewOffer(g *gin.Engine, offerHandler handler.Offer) *Offer {
	return &Offer{
		g:            g,
		offerHandler: offerHandler,
	}
}

func (r *Offer) Register(authMw middleware.Auth) {
	r.g.POST("/consumer/v1/offers", authMw.Consumer, r.offerHandler.ConsumerCreate)
	r.g.GET("/consumer/v1/offers", authMw.Consumer, r.offerHandler.ConsumerGetAll)
	r.g.GET("/consumer/v1/offers/:id", authMw.Consumer, r.offerHandler.ConsumerGetByID)

	r.g.POST("/provider/v1/offers/:id", authMw.ServiceProvider, r.offerHandler.ProviderAction)
	r.g.GET("/provider/v1/offers", authMw.ServiceProvider, r.offerHandler.ProviderGetAll)
	r.g.GET("/provider/v1/offers/:id", authMw.ServiceProvider, r.offerHandler.ProviderGetByID)
}
