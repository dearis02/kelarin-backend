package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

const ServiceElasticSearchIndexName = "services"

// region repo types

type Service struct {
	ID                    uuid.UUID       `db:"id"`
	ServiceProviderID     uuid.UUID       `db:"service_provider_id"`
	Name                  string          `db:"name"`
	Description           string          `db:"description"`
	DeliveryMethods       DeliveryMethods `db:"delivery_methods"`
	FeeStartAt            decimal.Decimal `db:"fee_start_at"`
	FeeEndAt              decimal.Decimal `db:"fee_end_at"`
	Rules                 ServiceRules    `db:"rules"`
	Images                pq.StringArray  `db:"images"`
	IsAvailable           bool            `db:"is_available"`
	ReceivedRatingCount   int32           `db:"received_rating_count"`
	ReceivedRatingAverage float32         `db:"received_rating_average"`
	IsDeleted             bool            `db:"is_deleted"`
	CreatedAt             time.Time       `db:"created_at"`
	DeletedAt             null.Time       `db:"deleted_at"`
}

type DeliveryMethods []ServiceDeliveryMethod

type ServiceDeliveryMethod string

func (t DeliveryMethods) Value() (driver.Value, error) {
	return pq.Array(t).Value()
}

func (t *DeliveryMethods) Scan(src any) error {
	if src == nil {
		return nil
	}

	source, ok := src.([]byte)
	if !ok {
		return errors.New("types.DeliveryMethods: invalid type")
	}

	return pq.Array(t).Scan(source)
}

func (t *ServiceDeliveryMethod) Scan(src any) error {
	if src == nil {
		return nil
	}

	source, ok := src.([]byte)
	if !ok {
		return errors.New("types.ServiceDeliveryMethod: invalid type")
	}

	*t = ServiceDeliveryMethod(string(source))

	return nil
}

const (
	ServiceDeliveryMethodOnsite ServiceDeliveryMethod = "onsite"
	ServiceDeliveryMethodOnline ServiceDeliveryMethod = "online"
)

type ServiceRule struct {
	Type ServiceRuleType `json:"type"`
	Name string          `json:"name"`
}

type ServiceRules []ServiceRule

func (t ServiceRules) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *ServiceRules) Scan(src any) error {
	if src == nil {
		return nil
	}

	source, ok := src.([]byte)
	if !ok {
		return errors.New("types.ServiceRules: invalid type")
	}

	return json.Unmarshal(source, t)
}

func (r ServiceRule) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required),
		validation.Field(&r.Type, validation.Required, validation.In(ServiceRuleTypeInclusion, ServiceRuleTypeExclusion)),
	)
}

type ServiceRuleType string

const (
	ServiceRuleTypeInclusion ServiceRuleType = "inclusion"
	ServiceRuleTypeExclusion ServiceRuleType = "exclusion"
)

// end of region repo types

// region service types

type ServiceCreateReq struct {
	AuthUser        AuthUser                `middleware:"user"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	DeliveryMethods []ServiceDeliveryMethod `json:"delivery_methods"`
	FeeStartAt      decimal.Decimal         `json:"fee_start_at"`
	FeeEndAt        decimal.Decimal         `json:"fee_end_at"`
	Rules           []ServiceRule           `json:"rules"`
	Images          []string                `json:"images"`
	IsAvailable     bool                    `json:"is_available"`
	CategoryIDs     []uuid.UUID             `json:"category_ids"`
}

func (r ServiceCreateReq) Validate() error {
	err := validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required),
		validation.Field(&r.Description, validation.Required),
		validation.Field(&r.DeliveryMethods, validation.Required, validation.Each(validation.In(ServiceDeliveryMethodOnsite, ServiceDeliveryMethodOnline))),
		validation.Field(&r.FeeStartAt, validation.Required),
		validation.Field(&r.FeeEndAt, validation.Required),
		validation.Field(&r.Rules, validation.Required),
		validation.Field(&r.Images, validation.Required),
		validation.Field(&r.CategoryIDs, validation.Required),
	)

	if err != nil {
		return err
	}

	ve := validation.Errors{}

	if r.FeeStartAt.LessThan(decimal.Zero) {
		ve["fee_start_at"] = validation.NewError("fee_start_at", "must be greater than or equal to 0")
	}

	if r.FeeEndAt.LessThan(decimal.Zero) {
		ve["fee_end_at"] = validation.NewError("fee_end_at", "must be greater than or equal to 0")
	}

	if r.FeeStartAt.GreaterThan(r.FeeEndAt) {
		ve["fee_start_at"] = validation.NewError("fee_start_at", "must be less than fee_end_to")
	}

	if len(ve) > 0 {
		return ve
	}

	return nil
}

type ServiceIndex struct {
	ID              uuid.UUID               `json:"id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Province        null.String             `json:"province"`
	City            null.String             `json:"city"`
	DeliveryMethods []ServiceDeliveryMethod `json:"delivery_methods"`
	Categories      []string                `json:"categories"`
	Rules           ServiceRules            `json:"rules"`
	FeeStartAt      decimal.Decimal         `json:"fee_start_at"`
	FeeEndAt        decimal.Decimal         `json:"fee_end_at"`
	IsAvailable     bool                    `json:"is_available"`
	CreatedAt       time.Time               `json:"created_at"`
}

type ServiceGetByIDReq struct {
	AuthUser AuthUser  `middleware:"user"`
	ID       uuid.UUID `uri:"id" binding:"required,uuid"`
}

type ServiceGetByIDRes struct {
	ID              uuid.UUID               `json:"id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	DeliveryMethods []ServiceDeliveryMethod `json:"delivery_methods"`
	Categories      []ServiceCategoryRes    `json:"categories"`
	FeeStartAt      decimal.Decimal         `json:"fee_start_at"`
	FeeEndAt        decimal.Decimal         `json:"fee_end_at"`
	Rules           []ServiceRule           `json:"rules"`
	Images          []ImageRes              `json:"images"`
	IsAvailable     bool                    `json:"is_available"`
	CreatedAt       time.Time               `json:"created_at"`
}

type ServiceUpdateReq struct {
	AuthUser        AuthUser                `middleware:"user"`
	ID              uuid.UUID               `uri:"id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	DeliveryMethods []ServiceDeliveryMethod `json:"delivery_methods"`
	FeeStartAt      decimal.Decimal         `json:"fee_start_at"`
	FeeEndAt        decimal.Decimal         `json:"fee_end_at"`
	Rules           []ServiceRule           `json:"rules"`
	IsAvailable     bool                    `json:"is_available"`
	CategoryIDs     []uuid.UUID             `json:"category_ids"`
}

func (r ServiceUpdateReq) Validate() error {
	err := validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required),
		validation.Field(&r.Description, validation.Required),
		validation.Field(&r.DeliveryMethods, validation.Required, validation.Each(validation.In(ServiceDeliveryMethodOnsite, ServiceDeliveryMethodOnline))),
		validation.Field(&r.FeeStartAt, validation.Required),
		validation.Field(&r.FeeEndAt, validation.Required),
		validation.Field(&r.Rules, validation.Required),
		validation.Field(&r.CategoryIDs, validation.Required),
	)

	if err != nil {
		return err
	}

	ve := validation.Errors{}

	if r.FeeStartAt.LessThan(decimal.Zero) {
		ve["fee_start_at"] = validation.NewError("fee_start_at", "must be greater than or equal to 0")
	}

	if r.FeeEndAt.LessThan(decimal.Zero) {
		ve["fee_end_at"] = validation.NewError("fee_end_at", "must be greater than or equal to 0")
	}

	if r.FeeStartAt.GreaterThan(r.FeeEndAt) {
		ve["fee_start_at"] = validation.NewError("fee_start_at", "must be less than fee_end_to")
	}

	if len(ve) > 0 {
		return ve
	}

	return nil
}

type ServiceGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
}

func (r ServiceGetAllReq) Validate() error {
	if r.AuthUser == (AuthUser{}) {
		return errors.New("AuthUser is required")
	}

	return nil
}

type ServiceGetAllRes struct {
	ID              uuid.UUID               `json:"id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	DeliveryMethods []ServiceDeliveryMethod `json:"delivery_methods"`
	FeeStartAt      decimal.Decimal         `json:"fee_start_at"`
	FeeEndAt        decimal.Decimal         `json:"fee_end_at"`
	Rules           ServiceRules            `json:"rules"`
	IsAvailable     bool                    `json:"is_available"`
	CreatedAt       time.Time               `json:"created_at"`
	Categories      []ServiceCategoryRes    `json:"categories"`
}

type ServiceDeleteReq struct {
	AuthUser AuthUser  `middleware:"user"`
	ID       uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r ServiceDeleteReq) Validate() error {
	if r.AuthUser == (AuthUser{}) {
		return errors.New("AuthUser is required")
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required),
	)
}

type ServiceImageActionReq struct {
	AuthUser  AuthUser  `middleware:"user"`
	ImageKeys []string  `json:"image_keys"`
	ID        uuid.UUID `uri:"id"`
}

func (r ServiceImageActionReq) Validate() error {
	if r.AuthUser == (AuthUser{}) {
		return errors.New("AuthUser is required")
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.ImageKeys, validation.Required),
	)
}

type ServiceIndexFilter struct {
	Limit           int
	LatestTimestamp null.Time
	Keyword         string
	Province        string
	City            string
	Categories      []string
}

// end of region service types

const ServiceImageDir = "images/service"
