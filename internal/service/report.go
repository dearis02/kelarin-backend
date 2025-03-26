package service

import (
	"context"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"kelarin/internal/utils"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-errors/errors"
)

type Report interface {
	ProviderGetMonthlySummary(ctx context.Context, req types.ReportProviderGetMonthlySummaryReq) (types.ReportProviderGetMonthlySummaryRes, error)
	ProviderExportOrders(ctx context.Context, req types.ReportProviderExportOrdersReq) (types.ReportProviderExportOrdersRes, error)
	ProviderExportMonthlySummary(ctx context.Context, req types.ReportProviderExportMonthlySummaryReq) (types.ReportProviderExportMonthlySummaryRes, error)
}

type reportImpl struct {
	serviceProviderRepo repository.ServiceProvider
	offerRepo           repository.Offer
	orderRepo           repository.Order
	utilSvc             Util
}

func NewReport(serviceProviderRepo repository.ServiceProvider, offerRepo repository.Offer, orderRepo repository.Order, utilSvc Util) Report {
	return &reportImpl{
		serviceProviderRepo: serviceProviderRepo,
		offerRepo:           offerRepo,
		orderRepo:           orderRepo,
		utilSvc:             utilSvc,
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
		var offerCount int64
		for _, o := range offerReports {
			if o.Date.Equal(d) {
				offerCount = o.Count
				break
			}
		}

		res.MonthlyReceivedOffers = append(res.MonthlyReceivedOffers, types.ReportProviderGetSummaryResMonthlyReceivedOffers{
			Date:  d.Day(),
			Total: offerCount,
		})

		var orderCount int64
		for _, o := range orderReports {
			if o.Date.Equal(d) {
				orderCount = o.Count
				break
			}
		}

		res.MonthlyReceivedOrders = append(res.MonthlyReceivedOrders, types.ReportProviderGetSummaryResMonthlyReceivedOrders{
			Date:  d.Day(),
			Total: orderCount,
		})
	}

	return res, nil
}

func (s *reportImpl) ProviderExportOrders(ctx context.Context, req types.ReportProviderExportOrdersReq) (types.ReportProviderExportOrdersRes, error) {
	res := types.ReportProviderExportOrdersRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.Errorf("service provider not found: user_id %s", req.AuthUser.ID)
	} else if err != nil {
		return res, err
	}

	orders, err := s.orderRepo.FindForReportExportByServiceProviderID(ctx, provider.ID)
	if err != nil {
		return res, err
	}

	fileName := s.generateReportFileName("orders")
	filePath := filepath.Join(types.TempFileDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return res, err
	}
	defer file.Close()

	reqTz, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	csvRows := []types.ReportProviderGetAllOrderCSV{}

	for _, o := range orders {
		paymentFulfilled := "No"
		if o.PaymentFulfilled {
			paymentFulfilled = "Yes"
		}

		csvRows = append(csvRows, types.ReportProviderGetAllOrderCSV{
			ID:               o.ID.String(),
			ServiceFee:       o.ServiceFee.String(),
			ServiceDate:      o.ServiceDate.Format(time.DateOnly),
			ServiceTime:      o.ServiceTime.In(reqTz).Format(time.TimeOnly),
			Status:           string(o.Status),
			PaymentFulfilled: paymentFulfilled,
			CustomerName:     o.UserName,
			CustomerEmail:    o.UserEmail,
			CustomerProvince: o.UserProvince,
			CustomerCity:     o.UserCity,
			CustomerAddress:  o.UserAddress,
			CreatedAt:        o.CreatedAt.Format(time.DateTime),
		})
	}

	err = utils.WriteCSV(csvRows, file)
	if errors.Is(err, types.ErrEmptySlice) {
		return res, errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "no data to export"})
	} else if err != nil {
		return res, err
	}

	res = types.ReportProviderExportOrdersRes{
		FileName: fileName,
		FilePath: filePath,
	}

	return res, nil
}

func (reportImpl) generateReportFileName(prefix string) string {
	timeNow := time.Now()
	fileName := fmt.Sprintf("%s_%s_%d.csv", prefix, timeNow.Format(time.DateTime), timeNow.Unix())

	return fileName
}

func (s *reportImpl) ProviderExportMonthlySummary(ctx context.Context, req types.ReportProviderExportMonthlySummaryReq) (types.ReportProviderExportMonthlySummaryRes, error) {
	res := types.ReportProviderExportMonthlySummaryRes{}

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

	offerGroupByStatusCount, err := s.offerRepo.CountGroupByStatusByServiceProviderIDAndMonthAndYear(ctx, provider.ID, req.Month, req.Year)
	if err != nil {
		return res, err
	}

	orderGroupByStatusCount, err := s.orderRepo.CountGroupByStatusByServiceProviderIDAndMonthAndYear(ctx, provider.ID, req.Month, req.Year)
	if err != nil {
		return res, err
	}

	totalServiceFee, err := s.orderRepo.SumServiceFeeByServiceProviderIDAndStatusAndMonthAndYear(ctx, provider.ID, types.OrderStatusFinished, req.Month, req.Year)
	if err != nil {
		return res, err
	}

	csv := []types.ReportProviderGetMonthlySummaryCSV{
		{
			PendingOfferCount:  "0",
			AcceptedOfferCount: "0",
			RejectedOfferCount: "0",
			CanceledOfferCount: "0",
			PendingOrderCount:  "0",
			OngoingOrderCount:  "0",
			FinishedOrderCount: "0",
			TotalIncome:        totalServiceFee.String(),
		},
	}

	for status, count := range offerGroupByStatusCount {
		switch status {
		case types.OfferStatusPending:
			csv[0].PendingOfferCount = strconv.FormatInt(count, 10)
		case types.OfferStatusAccepted:
			csv[0].AcceptedOfferCount = strconv.FormatInt(count, 10)
		case types.OfferStatusRejected:
			csv[0].RejectedOfferCount = strconv.FormatInt(count, 10)
		case types.OfferStatusCanceled:
			csv[0].CanceledOfferCount = strconv.FormatInt(count, 10)
		}
	}

	for status, count := range orderGroupByStatusCount {
		switch status {
		case types.OrderStatusPending:
			csv[0].PendingOrderCount = strconv.FormatInt(count, 10)
		case types.OrderStatusOngoing:
			csv[0].OngoingOrderCount = strconv.FormatInt(count, 10)
		case types.OrderStatusFinished:
			csv[0].FinishedOrderCount = strconv.FormatInt(count, 10)
		}
	}

	fileName := s.generateReportFileName("monthly_summary")
	filePath := filepath.Join(types.TempFileDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return res, errors.New(err)
	}

	err = utils.WriteCSV(csv, file)
	if errors.Is(err, types.ErrEmptySlice) {
		return res, errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "no data to export"})
	} else if err != nil {
		return res, err
	}

	res.FileName = fileName
	res.FilePath = filePath

	return res, nil
}
