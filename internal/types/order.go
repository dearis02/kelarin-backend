package types

import (
	dbUtil "kelarin/internal/utils/dbutil"
	"time"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

// region repo types

type Order struct {
	ID                uuid.UUID       `db:"id"`
	UserID            uuid.UUID       `db:"user_id"`
	ServiceProviderID uuid.UUID       `db:"service_provider_id"`
	OfferID           uuid.UUID       `db:"offer_id"`
	PaymentID         uuid.NullUUID   `db:"payment_id"`
	PaymentFulfilled  bool            `db:"payment_fulfilled"`
	ServiceFee        decimal.Decimal `db:"service_fee"`
	ServiceDate       time.Time       `db:"service_date"`
	ServiceTime       time.Time       `db:"service_time"`
	Status            OrderStatus     `db:"status"`
	CreatedAt         time.Time       `db:"created_at"`
	UpdatedAt         null.Time       `db:"updated_at"`
}

type OrderStatus string

const (
	OrderStatusPending  OrderStatus = "pending"
	OrderStatusOngoing  OrderStatus = "ongoing"
	OrderStatusFinished OrderStatus = "finished"
	OrderStatusExpired  OrderStatus = "expired"
)

type OrderWithRelations struct {
	Order
	ServiceID   uuid.UUID   `db:"service_id"`
	ServiceName string      `db:"service_name"`
	OfferStatus OfferStatus `db:"offer_status"`
	UserName    string      `db:"user_name"`
	UserEmail   string      `db:"user_email"`
}

type OrderWithUserAndServiceProvider struct {
	Order
	ServiceProviderUserID uuid.UUID `db:"service_provider_user_id"`
	ServiceProviderName   string    `db:"service_provider_name"`
	UserName              string    `db:"user_name"`
}

type OrderWithServiceAndServiceProvider struct {
	Order
	ServiceID                uuid.UUID           `db:"service_id"`
	ServiceName              string              `db:"service_name"`
	ServiceProviderName      string              `db:"service_provider_name"`
	ServiceProviderLogoImage string              `db:"service_provider_logo_image"`
	PaymentMethodName        null.String         `db:"payment_method_name"`
	PaymentStatus            null.String         `db:"payment_status"`
	PaymentAmount            decimal.NullDecimal `db:"payment_amount"`
	PaymentAdminFee          null.Int32          `db:"payment_admin_fee"`
	PaymentPlatformFee       null.Int32          `db:"payment_platform_fee"`
	PaymentPaymentLink       null.String         `db:"payment_payment_link"`
	PaymentCreatedAt         null.Time           `db:"payment_created_at"`
	PaymentExpiredAt         null.Time           `db:"payment_expired_at"`
}

type OrderForReport struct {
	Date  time.Time `db:"date"`
	Count int64     `db:"count"`
}

type OrderForReportExport struct {
	ID               uuid.UUID       `db:"id"`
	ServiceFee       decimal.Decimal `db:"service_fee"`
	ServiceDate      time.Time       `db:"service_date"`
	ServiceTime      time.Time       `db:"service_time"`
	Status           OrderStatus     `db:"status"`
	PaymentFulfilled bool            `db:"payment_fulfilled"`
	UserName         string          `db:"user_name"`
	UserEmail        string          `db:"user_email"`
	UserProvince     string          `db:"user_province"`
	UserCity         string          `db:"user_city"`
	UserAddress      string          `db:"user_address"`
	CreatedAt        time.Time       `db:"created_at"`
}

// endregion repo types

// region service types

type OrderCreateReq struct {
	AuthUser AuthUser
	Offer
	UserAddress
	ServiceProviderID      uuid.UUID
	ServiceName            string
	ServiceDeliveryMethods DeliveryMethods
	ServiceRules           []ServiceRule
	ServiceDescription     string
	ServiceDate            time.Time
	ServiceTime            time.Time
	Tx                     dbUtil.Tx
}

func (r OrderCreateReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.Offer.ID == uuid.Nil {
		return errors.New("Offer.ID is required")
	}

	if r.Offer.UserID == uuid.Nil {
		return errors.New("Offer.UserID is required")
	}

	if r.ServiceProviderID == uuid.Nil {
		return errors.New("ServiceProviderID is required")
	}

	if r.ServiceDate.IsZero() {
		return errors.New("ServiceDate is required")
	}

	if r.ServiceTime.IsZero() {
		return errors.New("ServiceTime is required")
	}

	if r.Tx == nil {
		return errors.New("Tx is required")
	}

	return nil
}

type OrderConsumerGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
	TimeZone string   `header:"Time-Zone"`
}

func (r OrderConsumerGetAllReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return nil
}

type OrderConsumerGetAllRes struct {
	ID               uuid.UUID                             `json:"id"`
	OfferID          uuid.UUID                             `json:"offer_id"`
	ServiceFee       decimal.Decimal                       `json:"service_fee"`
	ServiceDate      string                                `json:"service_date"`
	ServiceTime      string                                `json:"service_time"`
	PaymentFulfilled bool                                  `json:"payment_fulfilled"`
	Status           OrderStatus                           `json:"status"`
	CreatedAt        time.Time                             `json:"created_at"`
	Service          OrderConsumerGetAllResService         `json:"service"`
	ServiceProvider  OrderConsumerGetAllResServiceProvider `json:"service_provider"`
	Payment          *OrderConsumerGetAllResPayment        `json:"payment"`
}

type OrderConsumerGetAllResService struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type OrderConsumerGetAllResServiceProvider struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	LogoURL string    `json:"logo_url"`
}

type OrderConsumerGetAllResPayment struct {
	ID                uuid.UUID       `json:"id"`
	PaymentMethodName string          `json:"payment_method_name"`
	Amount            decimal.Decimal `json:"amount"`
	AdminFee          int32           `json:"admin_fee"`
	PlatformFee       int32           `json:"platform_fee"`
	Status            PaymentStatus   `json:"status"`
	PaymentLink       string          `json:"payment_link"`
	CreatedAt         time.Time       `json:"created_at"`
	ExpiredAt         time.Time       `json:"expired_at"`
}

type OrderConsumerGetByIDReq struct {
	AuthUser AuthUser  `middleware:"user"`
	ID       uuid.UUID `param:"id"`
	TimeZone string    `header:"Time-Zone" `
}

func (r OrderConsumerGetByIDReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.ID == uuid.Nil {
		return errors.New(ErrIDRouteParamRequired)
	}

	return nil
}

type ConsumerOrderGetByIDRes struct {
	ID               uuid.UUID                           `json:"id"`
	OfferID          uuid.UUID                           `json:"offer_id"`
	ServiceFee       decimal.Decimal                     `json:"service_fee"`
	ServiceDate      string                              `json:"service_date"`
	ServiceTime      string                              `json:"service_time"`
	PaymentFulfilled bool                                `json:"payment_fulfilled"`
	Status           OrderStatus                         `json:"status"`
	Rated            bool                                `json:"rated"`
	CreatedAt        time.Time                           `json:"created_at"`
	Offer            ConsumerOrderGetByIDResOffer        `json:"offer"`
	Service          ConsumerOrderGetByIDResOfferService `json:"service"`
	Address          ConsumerOrderGetByIDResOfferAddress `json:"address"`
	Payment          *OrderConsumerGetByIDResPayment     `json:"payment"`
}

type ConsumerOrderGetByIDResOffer struct {
	ID        uuid.UUID   `json:"id"`
	Detail    string      `json:"detail"`
	Status    OfferStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type ConsumerOrderGetByIDResOfferService struct {
	ID              uuid.UUID                                     `json:"id"`
	Name            string                                        `json:"name"`
	DeliveryMethods DeliveryMethods                               `json:"delivery_methods"`
	Rules           []ServiceRule                                 `json:"rules"`
	Description     string                                        `json:"description"`
	ServiceProvider ConsumerOrderGetByIDResServiceServiceProvider `json:"service_provider"`
}

type ConsumerOrderGetByIDResServiceServiceProvider struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	LogoURL               string    `json:"logo_url"`
	ReceivedRatingCount   int32     `json:"received_rating_count"`
	ReceivedRatingAverage float64   `json:"received_rating_average"`
}

type ConsumerOrderGetByIDResOfferAddress struct {
	Province string       `json:"province"`
	City     string       `json:"city"`
	Lat      null.Float64 `json:"lat"`
	Lng      null.Float64 `json:"lng"`
	Detail   string       `json:"detail"`
}

type OrderConsumerGetByIDResPayment struct {
	ID                uuid.UUID       `json:"id"`
	Reference         string          `json:"reference"`
	PaymentMethodName string          `json:"payment_method_name"`
	PaymentMethodLogo string          `json:"payment_method_logo"`
	Amount            decimal.Decimal `json:"amount"`
	AdminFee          int32           `json:"admin_fee"`
	PlatformFee       int32           `json:"platform_fee"`
	Status            PaymentStatus   `json:"status"`
	PaymentLink       string          `json:"payment_link"`
	ExpiredAt         time.Time       `json:"expired_at"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         null.Time       `json:"updated_at"`
}

type OrderConsumerGenerateQRCodeReq struct {
	AuthUser AuthUser  `middleware:"user"`
	ID       uuid.UUID `param:"id"`
}

func (r OrderConsumerGenerateQRCodeReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.ID == uuid.Nil {
		return errors.New(ErrIDRouteParamRequired)
	}

	return nil
}

type OrderConsumerGenerateQRCodeRes struct {
	QRCodeContent         string  `json:"qr_code_content"`
	ValidDurationInSecond float64 `json:"valid_duration_in_second"`
}

type OrderConsumerGenerateQRCodePayload struct {
	jwt.RegisteredClaims
	OrderID     uuid.UUID       `json:"order_id"`
	Amount      decimal.Decimal `json:"amount"`
	AdminFee    int32           `json:"admin_fee"`
	PlatformFee int32           `json:"platform_fee"`
}

type OrderProviderGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
	TimeZone string   `header:"Time-Zone"`
}

func (r OrderProviderGetAllReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return nil
}

type OrderProviderGetAllRes struct {
	ID               uuid.UUID                      `json:"id"`
	OfferID          uuid.UUID                      `json:"offer_id"`
	ServiceFee       decimal.Decimal                `json:"service_fee"`
	ServiceDate      string                         `json:"service_date"`
	ServiceTime      string                         `json:"service_time"`
	PaymentFulfilled bool                           `json:"payment_fulfilled"`
	Status           OrderStatus                    `json:"status"`
	CreatedAt        time.Time                      `json:"created_at"`
	Payment          *OrderProviderGetAllResPayment `json:"payment"`
}

type OrderProviderGetAllResPayment struct {
	ID                uuid.UUID       `json:"id"`
	PaymentMethodName string          `json:"payment_method_name"`
	Amount            decimal.Decimal `json:"amount"`
	AdminFee          int32           `json:"admin_fee"`
	PlatformFee       int32           `json:"platform_fee"`
	Status            PaymentStatus   `json:"status"`
}

type OrderProviderGetByIDReq struct {
	AuthUser AuthUser  `middleware:"user"`
	ID       uuid.UUID `param:"id"`
}

func (r OrderProviderGetByIDReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.ID == uuid.Nil {
		return errors.New(ErrIDRouteParamRequired)
	}

	return nil
}

type OrderProviderGetByIDRes struct {
	ID               uuid.UUID                      `json:"id"`
	OfferID          uuid.UUID                      `json:"offer_id"`
	ServiceName      string                         `json:"service_name"`
	ServiceFee       decimal.Decimal                `json:"service_fee"`
	ServiceDate      string                         `json:"service_date"`
	ServiceTime      string                         `json:"service_time"`
	PaymentFulfilled bool                           `json:"payment_fulfilled"`
	Status           OrderStatus                    `json:"status"`
	CreatedAt        time.Time                      `json:"created_at"`
	User             OrderProviderGetByIDResUser    `json:"user"`
	Offer            OrderProviderGetByIDResOffer   `json:"offer"`
	Address          OrderProviderGetByIDResAddress `json:"address"`
	Payment          *OrderProviderGetAllResPayment `json:"payment"`
}

type OrderProviderGetByIDResOffer struct {
	ID     uuid.UUID `json:"id"`
	Detail string    `json:"detail"`
}

type OrderProviderGetByIDResUser struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type OrderProviderGetByIDResAddress struct {
	Province string       `json:"province"`
	City     string       `json:"city"`
	Lat      null.Float64 `json:"latitude"`
	Lng      null.Float64 `json:"longitude"`
	Detail   string       `json:"detail"`
}

type OrderProviderValidateQRCodeReq struct {
	AuthUser      AuthUser `middleware:"user"`
	QRCodeContent string   `json:"qr_code_content"`
}

func (r OrderProviderValidateQRCodeReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.QRCodeContent, validation.Required),
	)
}

// endregion service types
