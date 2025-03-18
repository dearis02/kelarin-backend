package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"kelarin/internal/utils"
	"time"

	"github.com/go-errors/errors"
)

type Report interface {
	ProviderGetMonthlySummary(ctx context.Context, req types.ReportProviderGetMonthlySummaryReq) (types.ReportProviderGetMonthlySummaryRes, error)
}

type reportImpl struct {
	serviceProviderRepo repository.ServiceProvider
	offerRepo           repository.Offer
	orderRepo           repository.Order
}

func NewReport(serviceProviderRepo repository.ServiceProvider, offerRepo repository.Offer, orderRepo repository.Order) Report {
	return &reportImpl{
		serviceProviderRepo: serviceProviderRepo,
		offerRepo:           offerRepo,
		orderRepo:           orderRepo,
	}
}

func (s *reportImpl) ProviderGetMonthlySummary(ctx context.Context, req types.ReportProviderGetMonthlySummaryReq) (types.ReportProviderGetMonthlySummaryRes, error) {
	res := types.ReportProviderGetMonthlySummaryRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	req.SetDefaultMonthAndYear()

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("service provider not found: user_id %s", req.AuthUser.ID)
	} else if err != nil {
		return res, err
	}

	totalOffers, offerReports, err := s.offerRepo.FindForReportByServiceProviderID(ctx, provider.ID, req.Month, req.Year)
	if err != nil {
		return res, err
	}

	totalOrders, orderReports, err := s.orderRepo.FindForReportByServiceProviderID(ctx, provider.ID, req.Month, req.Year)
	if err != nil {
		return res, err
	}

	totalServiceFee, err := s.orderRepo.FindTotalServiceFeeByServiceProviderIDAndStatusAndMonthAndYear(ctx, provider.ID, types.OrderStatusFinished, req.Month, req.Year)
	if err != nil {
		return res, err
	}

	days := utils.GenerateDaysInMonth(req.Year, time.Month(req.Month))

	res.Month = req.Month
	res.Year = req.Year
	res.MonthlyTotalIncome = totalServiceFee
	res.MonthlyTotalReceivedOffers = totalOffers
	res.MonthlyTotalReceivedOrders = totalOrders

	for _, d := range days {
		var count int64
		for _, o := range offerReports {
			if o.Date.Equal(d) {
				count = o.Count
				break
			}
		}

		res.MonthlyReceivedOffers = append(res.MonthlyReceivedOffers, types.ReportProviderGetSummaryResMonthlyReceivedOffers{
			Date:  d.Day(),
			Total: count,
		})

		for _, o := range orderReports {
			if o.Date.Equal(d) {
				count = o.Count
				break
			}
		}

		res.MonthlyReceivedOrders = append(res.MonthlyReceivedOrders, types.ReportProviderGetSummaryResMonthlyReceivedOrders{
			Date:  d.Day(),
			Total: count,
		})
	}

	return res, nil
}
