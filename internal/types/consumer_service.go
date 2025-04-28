package types

import (
	"net/http"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// region service types

type ConsumerServiceGetAllReq struct {
	Province   string
	City       string
	Categories []string
	Keyword    string
	After      string
	PaginationReq
}

type ConsumerServiceGetAllRes struct {
	ID                    uuid.UUID       `json:"id"`
	Name                  string          `json:"name"`
	ImageURL              string          `json:"image_url"`
	FeeStartAt            decimal.Decimal `json:"fee_start_at"`
	FeeEndAt              decimal.Decimal `json:"fee_end_at"`
	Address               string          `json:"address"`
	Province              string          `json:"province"`
	City                  string          `json:"city"`
	ReceivedRatingCount   int32           `json:"received_rating_count"`
	ReceivedRatingAverage float32         `json:"received_rating_average"`
}

type ConsumerServiceGetByIDRes struct {
	ID                    uuid.UUID                         `json:"id"`
	Name                  string                            `json:"name"`
	Description           string                            `json:"description"`
	DeliveryMethods       DeliveryMethods                   `json:"delivery_methods"`
	ImageURLs             []string                          `json:"image_urls"`
	FeeStartAt            decimal.Decimal                   `json:"fee_start_at"`
	FeeEndAt              decimal.Decimal                   `json:"fee_end_at"`
	Rules                 ServiceRules                      `json:"rules"`
	IsAvailable           bool                              `json:"is_available"`
	ReceivedRatingCount   int32                             `json:"received_rating_count"`
	ReceivedRatingAverage float32                           `json:"received_rating_average"`
	ServiceProvider       ConsumerServiceServiceProviderRes `json:"service_provider"`
}

type ConsumerServiceServiceProviderRes struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	Province              string    `json:"province"`
	City                  string    `json:"city"`
	Address               string    `json:"address"`
	MobilePhoneNumber     string    `json:"mobile_phone_number"`
	Telephone             string    `json:"telephone"`
	LogoImageURL          string    `json:"logo_image_url"`
	ReceivedRatingCount   int32     `json:"received_rating_count"`
	ReceivedRatingAverage float64   `json:"received_rating_average"`
	JoinedAt              string    `json:"joined_at"`
}

type ConsumerServiceFeedbackCreateReq struct {
	AuthUser AuthUser  `middleware:"user"`
	OrderID  uuid.UUID `json:"order_id"`
	Rating   int16     `json:"rating"`
	Comment  string    `json:"comment"`
}

func (r ConsumerServiceFeedbackCreateReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.OrderID == uuid.Nil {
		return errors.New(AppErr{
			Code:    http.StatusUnprocessableEntity,
			Message: "order_id is required",
		})
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.OrderID, validation.Required),
		validation.Field(&r.Rating, validation.Required, validation.Min(1), validation.Max(5)),
	)
}

// end of region service types
