package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type UserAddress struct {
	g                  *gin.Engine
	userAddressHandler handler.UserAddress
}

func NewUserAddress(g *gin.Engine, userAddressHandler handler.UserAddress) *UserAddress {
	return &UserAddress{
		g:                  g,
		userAddressHandler: userAddressHandler,
	}
}

func (r *UserAddress) Register(authMw middleware.Auth) {
	r.g.GET("/consumer/v1/addresses", authMw.Consumer, r.userAddressHandler.GetAll)
	r.g.POST("/consumer/v1/addresses", authMw.Consumer, r.userAddressHandler.Create)
	r.g.PUT("/consumer/v1/addresses/:id", authMw.Consumer, r.userAddressHandler.Update)
}
