package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type ServiceCategory struct {
	g                      *gin.Engine
	serviceCategoryHandler handler.ServiceCategory
}

func NewServiceCategory(g *gin.Engine, serviceCategoryHandler handler.ServiceCategory) ServiceCategory {
	return ServiceCategory{
		g:                      g,
		serviceCategoryHandler: serviceCategoryHandler,
	}
}

func (r *ServiceCategory) Register(mw middleware.Auth) {
	r.g.GET("/common/v1/service-categories", mw.Authenticated, r.serviceCategoryHandler.GetAll)
}
