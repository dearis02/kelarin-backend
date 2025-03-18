package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Report struct {
	g             *gin.Engine
	reportHandler handler.Report
}

func NewReport(g *gin.Engine, reportHandler handler.Report) *Report {
	return &Report{
		g:             g,
		reportHandler: reportHandler,
	}
}

func (r *Report) Register(m middleware.Auth) {
	r.g.GET("/provider/v1/reports/_monthly", m.ServiceProvider, r.reportHandler.ProviderGetMonthlySummary)
	r.g.GET("/provider/v1/report/orders", m.ServiceProvider, r.reportHandler.ProviderExportOrders)
}
