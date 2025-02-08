package routes

import (
	"kelarin/internal/handler"

	"github.com/gin-gonic/gin"
)

type Province struct {
	g               *gin.Engine
	provinceHandler handler.Province
}

func NewProvince(g *gin.Engine, provinceHandler handler.Province) Province {
	return Province{
		g:               g,
		provinceHandler: provinceHandler,
	}
}

func (r *Province) Register() {
	r.g.GET("/common/v1/provinces", r.provinceHandler.GetAll)
}
