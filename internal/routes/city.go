package routes

import (
	"kelarin/internal/handler"

	"github.com/gin-gonic/gin"
)

type City struct {
	g       *gin.Engine
	cityHdl handler.City
}

func NewCity(g *gin.Engine, cityHdl handler.City) City {
	return City{
		g:       g,
		cityHdl: cityHdl,
	}
}

func (r *City) Register() {
	r.g.GET("/common/v1/cities", r.cityHdl.GetByProvinceID)
}
