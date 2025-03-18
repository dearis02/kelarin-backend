package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Report interface {
	ProviderGetMonthlySummary(c *gin.Context)
	ProviderExportOrders(c *gin.Context)
}

type reportImpl struct {
	reportService  service.Report
	authMiddleware middleware.Auth
}

func NewReport(reportService service.Report, authMiddleware middleware.Auth) Report {
	return &reportImpl{
		reportService:  reportService,
		authMiddleware: authMiddleware,
	}
}

func (h *reportImpl) ProviderGetMonthlySummary(c *gin.Context) {
	var req types.ReportProviderGetMonthlySummaryReq

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.reportService.ProviderGetMonthlySummary(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *reportImpl) ProviderExportOrders(c *gin.Context) {
	var req types.ReportProviderExportOrdersReq

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.reportService.ProviderExportOrders(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.FileAttachment(res.FilePath, res.FileName)
}
