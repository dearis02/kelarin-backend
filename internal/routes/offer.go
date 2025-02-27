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
}
