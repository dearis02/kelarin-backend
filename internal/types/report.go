package types

import (
	"time"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/shopspring/decimal"
)

// region service types

type ReportProviderGetMonthlySummaryReq struct {
	AuthUser AuthUser `middleware:"user"`
	Year     int      `form:"year" json:"year"`
	Month    int      `form:"month" json:"month"`
}

func (r ReportProviderGetMonthlySummaryReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.Month, validation.In(
			int(time.January),
			int(time.February),
			int(time.March),
			int(time.April),
			int(time.May),
			int(time.June),
			int(time.July),
			int(time.August),
			int(time.September),
			int(time.October),
			int(time.November),
			int(time.December),
		)),
	)
}

func (r *ReportProviderGetMonthlySummaryReq) SetDefaultMonthAndYear() {
	now := time.Now()

	if r.Month == 0 {
		r.Month = int(now.Month())
	}

	if r.Year == 0 {
		r.Year = now.Year()
	}
}

type ReportProviderGetMonthlySummaryRes struct {
	Month                      int                                                `json:"month"`
	Year                       int                                                `json:"year"`
	MonthlyTotalIncome         decimal.Decimal                                    `json:"monthly_total_income"`
	MonthlyTotalReceivedOffers int64                                              `json:"monthly_total_received_offers"`
	MonthlyTotalReceivedOrders int64                                              `json:"monthly_total_received_orders"`
	MonthlyReceivedOffers      []ReportProviderGetSummaryResMonthlyReceivedOffers `json:"monthly_received_offers"`
	MonthlyReceivedOrders      []ReportProviderGetSummaryResMonthlyReceivedOrders `json:"monthly_received_orders"`
}

type ReportProviderGetSummaryResMonthlyReceivedOffers struct {
	Date  int   `json:"date"`
	Total int64 `json:"total"`
}

type ReportProviderGetSummaryResMonthlyReceivedOrders struct {
	Date  int   `json:"date"`
	Total int64 `json:"total"`
}

// endregion service types
