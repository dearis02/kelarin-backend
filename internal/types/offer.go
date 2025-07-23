package types

import (
	"fmt"
	"kelarin/internal/utils"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

// region repo types

type Offer struct {
	ID               uuid.UUID       `db:"id"`
	UserID           uuid.UUID       `db:"user_id"`
	UserAddressID    uuid.UUID       `db:"user_address_id"`
	ServiceID        uuid.UUID       `db:"service_id"`
	Detail           string          `db:"detail"`
	ServiceCost      decimal.Decimal `db:"service_cost"`
	ServiceStartDate time.Time       `db:"service_start_date"`
	ServiceEndDate   time.Time       `db:"service_end_date"`
	ServiceStartTime time.Time       `db:"service_start_time"`
	ServiceEndTime   time.Time       `db:"service_end_time"`
	Status           OfferStatus     `db:"status"`
	CreatedAt        time.Time       `db:"created_at"`
}

type OfferStatus string

const (
	OfferStatusPending  OfferStatus = "pending"
	OfferStatusAccepted OfferStatus = "accepted"
	OfferStatusRejected OfferStatus = "rejected"
	OfferStatusCanceled OfferStatus = "canceled"
	OfferStatusExpired  OfferStatus = "expired"
)

type OfferWithServiceAndProvider struct {
	Offer
	ServiceName         string    `db:"service_name"`
	ServiceImage        string    `db:"service_image"`
	ServiceProviderID   uuid.UUID `db:"service_provider_id"`
	ServiceProviderName string    `db:"service_provider_name"`
	ServiceProviderLogo string    `db:"service_provider_logo_image"`
}

type OfferForReport struct {
	Date  time.Time `db:"date"`
	Count int64     `db:"count"`
}

// endregion repo types

// region service types

type OfferConsumerCreateReq struct {
	AuthUser         AuthUser  `middleware:"user"`
	TimeZone         string    `header:"Time-Zone"`
	ServiceID        uuid.UUID `json:"service_id"`
	AddressID        uuid.UUID `json:"address_id"`
	Detail           string    `json:"detail"`
	ServiceCost      float64   `json:"service_cost"`
	ServiceStartDate string    `json:"service_start_date"`
	ServiceEndDate   string    `json:"service_end_date"`
	ServiceStartTime string    `json:"service_start_time"`
	ServiceEndTime   string    `json:"service_end_time"`
}

func (r OfferConsumerCreateReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	err := validation.ValidateStruct(&r,
		validation.Field(&r.ServiceID, validation.Required),
		validation.Field(&r.AddressID, validation.Required),
		validation.Field(&r.Detail, validation.Required),
		validation.Field(&r.ServiceCost, validation.Required),
		validation.Field(&r.ServiceStartDate, validation.Required, validation.Date(time.DateOnly)),
		validation.Field(&r.ServiceEndDate, validation.Required, validation.Date(time.DateOnly)),
		validation.Field(&r.ServiceStartTime, validation.Required, validation.Date(time.TimeOnly)),
		validation.Field(&r.ServiceEndTime, validation.Required, validation.Date(time.TimeOnly)),
	)

	if err != nil {
		return err
	}

	ve := validation.Errors{}

	startDate, err := time.Parse(time.DateOnly, r.ServiceStartDate)
	if err != nil {
		return errors.New(err)
	}

	endDate, err := time.Parse(time.DateOnly, r.ServiceEndDate)
	if err != nil {
		return errors.New(err)
	}

	today := utils.DateNowInUTC()
	if startDate.Before(today) {
		ve["service_start_date"] = validation.NewError("service_start_date_min", "service_start_date must be equal or greater than today")
	}

	if endDate.Before(startDate) {
		ve["service_end_date"] = validation.NewError("service_end_date_min", "service_end_date must be equal or greater than service_start_date")
	}

	if len(ve) > 0 {
		return ve
	}

	return nil
}

func (r OfferConsumerCreateReq) ValidateDateTimeAndServiceFee(userTz *time.Location, serviceFeeStartAt decimal.Decimal) error {
	ve := validation.Errors{}

	nowUTC := utils.DateNowInUTC()

	startDate, err := time.Parse(time.DateOnly, r.ServiceStartDate)
	if err != nil {
		return errors.New(err)
	}

	endDate, err := time.Parse(time.DateOnly, r.ServiceEndDate)
	if err != nil {
		return errors.New(err)
	}

	if endDate.Before(startDate) {
		ve["service_end_date"] = validation.NewError("service_end_date_min", "service_end_date must be equal or greater than service_start_date")
	}

	startTime, err := utils.ParseTimeString(r.ServiceStartTime, userTz)
	if err != nil {
		return err
	}

	endTime, err := utils.ParseTimeString(r.ServiceEndTime, userTz)
	if err != nil {
		return err
	}

	if startDate.Equal(nowUTC) {
		if startTime.Before(nowUTC.Truncate(time.Hour)) {
			ve["service_start_time"] = validation.NewError("service_start_time_min", "Start time at least 1 hour greater than now")
		}
	}

	if endDate.Equal(startDate) && endTime.Before(startTime) {
		ve["service_end_time"] = validation.NewError("service_end_time_min", "End time must be equal or greater than Start time")
	}

	serviceCost := decimal.NewFromFloat(r.ServiceCost)
	if serviceCost.LessThan(serviceFeeStartAt) {
		ve["service_cost"] = validation.NewError("service_cost_min", fmt.Sprintf("service_cost must be greater or equal than %s", serviceFeeStartAt))
	}

	if len(ve) > 0 {
		return ve
	}

	return nil
}

type OfferConsumerGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
	TimeZone string   `header:"Time-Zone"`
}

func (r OfferConsumerGetAllReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return nil
}

type OfferConsumerGetAllRes struct {
	ID                    uuid.UUID                             `json:"id"`
	ServiceCost           decimal.Decimal                       `json:"service_cost"`
	ServiceStartDate      string                                `json:"service_start_date"`
	ServiceEndDate        string                                `json:"service_end_date"`
	ServiceStartTime      string                                `json:"service_start_time"`
	ServiceEndTime        string                                `json:"service_end_time"`
	ServiceTimeTimeZone   string                                `json:"service_time_time_zone"`
	Status                OfferStatus                           `json:"status"`
	HasPendingNegotiation bool                                  `json:"has_pending_negotiation"`
	CreatedAt             time.Time                             `json:"created_at"`
	Service               OfferConsumerGetAllResService         `json:"service"`
	ServiceProvider       OfferConsumerGetAllResServiceProvider `json:"service_provider"`
}

type OfferConsumerGetAllResService struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	ImageURL string    `json:"image_url"`
}

type OfferConsumerGetAllResServiceProvider struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	LogoURL string    `json:"logo_url"`
}

type OfferConsumerGetByIDReq struct {
	AuthUser AuthUser  `middleware:"user"`
	ID       uuid.UUID `param:"id"`
	TimeZone string    `header:"Time-Zone"`
}

func (r OfferConsumerGetByIDReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.ID == uuid.Nil {
		return ErrIDRouteParamRequired
	}

	return nil
}

type OfferConsumerGetByIDRes struct {
	ID                    uuid.UUID                              `json:"id"`
	ServiceCost           decimal.Decimal                        `json:"service_cost"`
	Detail                string                                 `json:"detail"`
	ServiceStartDate      string                                 `json:"service_start_date"`
	ServiceEndDate        string                                 `json:"service_end_date"`
	ServiceStartTime      string                                 `json:"service_start_time"`
	ServiceEndTime        string                                 `json:"service_end_time"`
	ServiceTimeTimeZone   string                                 `json:"service_time_time_zone"`
	Status                OfferStatus                            `json:"status"`
	HasPendingNegotiation bool                                   `json:"has_pending_negotiation"`
	CreatedAt             time.Time                              `json:"created_at"`
	Service               OfferConsumerGetByIDResService         `json:"service"`
	ServiceProvider       OfferConsumerGetByIDResServiceProvider `json:"service_provider"`
	Address               OfferConsumerGetByIDResAddress         `json:"address"`
	Negotiations          []OfferConsumerGetByIDResNegotiation   `json:"negotiations"`
}

type OfferConsumerGetByIDResService struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type OfferConsumerGetByIDResServiceProvider struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	LogoURL               string    `json:"logo_url"`
	ReceivedRatingCount   int32     `json:"received_rating_count"`
	ReceivedRatingAverage float64   `json:"received_rating_average"`
}

type OfferConsumerGetByIDResAddress struct {
	ID       uuid.UUID    `json:"id"`
	Name     string       `json:"name"`
	Province string       `json:"province"`
	City     string       `json:"city"`
	Lat      null.Float64 `json:"lat"`
	Lng      null.Float64 `json:"lng"`
	Address  string       `json:"address"`
}

type OfferConsumerGetByIDResNegotiation struct {
	ID                   uuid.UUID              `json:"id"`
	Message              string                 `json:"message"`
	RequestedServiceCost decimal.Decimal        `json:"requested_service_cost"`
	Status               OfferNegotiationStatus `json:"status"`
	CreatedAt            time.Time              `json:"created_at"`
}

type OfferProviderActionReq struct {
	AuthUser AuthUser                     `middleware:"user"`
	TimeZone string                       `header:"Time-Zone"`
	ID       uuid.UUID                    `param:"id"`
	Action   OfferProviderActionReqAction `json:"action"`
	Date     string                       `json:"date"`
	Time     string                       `json:"time"`
}

type OfferProviderActionReqAction string

const (
	OfferProviderActionReqActionAccept OfferProviderActionReqAction = "accept"
	OfferProviderActionReqActionReject OfferProviderActionReqAction = "reject"
)

func (r OfferProviderActionReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.ID == uuid.Nil {
		return errors.New(AppErr{Code: http.StatusBadRequest, Message: ErrIDRouteParamRequired.Error()})
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.Action, validation.Required, validation.In(OfferProviderActionReqActionAccept, OfferProviderActionReqActionReject)),
		validation.Field(&r.Date,
			validation.Required.When(r.Action == OfferProviderActionReqActionAccept),
			validation.Date(time.DateOnly),
		),
		validation.Field(&r.Time,
			validation.Required.When(r.Action == OfferProviderActionReqActionAccept),
			validation.Date(time.TimeOnly),
		),
	)
}

func (r OfferProviderActionReq) ValidateDateAndTime(startDate, endDate time.Time, startTime, endTime string) error {
	ve := validation.Errors{}

	if r.Action == OfferProviderActionReqActionAccept {
		t, err := utils.IsDateBetween(r.Date, startDate, endDate, time.DateOnly)
		if err != nil {
			return err
		}

		if !t {
			ve["date"] = validation.NewError("date_min_max", fmt.Sprintf("date must be between %s - %s", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)))
		}

		userTz, err := time.LoadLocation(r.TimeZone)
		if err != nil {
			return errors.New(err)
		}

		localTz, err := time.LoadLocation("Asia/Makassar")
		if err != nil {
			return errors.New(err)
		}

		targetTime, err := utils.ParseTimeString(r.Time, userTz)
		if err != nil {
			return errors.New(err)
		}

		startT, err := utils.ParseTimeString(startTime, localTz)
		if err != nil {
			return errors.New(err)
		}

		endT, err := utils.ParseTimeString(endTime, localTz)
		if err != nil {
			return errors.New(err)
		}

		// add one day to endT if startT is greater than endT
		// case: start_time = 23:00, end_time = 01:00
		if endT.Before(startT) {
			endT = endT.AddDate(0, 0, 1)
		}

		t, err = utils.IsTimeBetween(targetTime, startT, endT)
		if err != nil {
			return err
		}

		if !t {
			ve["time"] = validation.NewError("time_min_max", fmt.Sprintf("time must be between %s - %s", startT.In(userTz).Format(time.TimeOnly), endT.In(userTz).Format(time.TimeOnly)))
		}
	}

	if len(ve) > 0 {
		return ve
	}

	return nil
}

type OfferProviderGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
}

func (r OfferProviderGetAllReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return nil
}

type OfferProviderGetAllRes struct {
	ID               uuid.UUID       `json:"id"`
	Detail           string          `json:"detail"`
	ServiceCost      decimal.Decimal `json:"service_cost"`
	ServiceStartDate string          `json:"service_start_date"`
	ServiceEndDate   string          `json:"service_end_date"`
	ServiceStartTime string          `json:"service_start_time"`
	ServiceEndTime   string          `json:"service_end_time"`
	Status           OfferStatus     `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
}

type OfferProviderGetByIDReq struct {
	AuthUser AuthUser  `middleware:"user"`
	ID       uuid.UUID `param:"id"`
}

func (r OfferProviderGetByIDReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.ID == uuid.Nil {
		return errors.New(ErrIDRouteParamRequired)
	}

	return nil
}

type OfferProviderGetByIDRes struct {
	OfferProviderGetAllRes
	Service      OfferProviderGetByIDResService       `json:"service"`
	User         OfferGetByIDResUser                  `json:"user"`
	Negotiations []OfferConsumerGetByIDResNegotiation `json:"negotiations"`
}

type OfferProviderGetByIDResService struct {
	ID         uuid.UUID       `json:"id"`
	Name       string          `json:"name"`
	FeeStartAt decimal.Decimal `json:"fee_start_at"`
	FeeEndAt   decimal.Decimal `json:"fee_end_at"`
}

type OfferGetByIDResUser struct {
	ID      uuid.UUID                  `json:"id"`
	Name    string                     `json:"name"`
	Address OfferGetByIDResUserAddress `json:"address"`
}

type OfferGetByIDResUserAddress struct {
	ID       uuid.UUID    `json:"id"`
	Name     string       `json:"name"`
	Lat      null.Float64 `json:"lat"`
	Lng      null.Float64 `json:"lng"`
	Province string       `json:"province"`
	City     string       `json:"city"`
	Address  string       `json:"address"`
}

// endregion service types
