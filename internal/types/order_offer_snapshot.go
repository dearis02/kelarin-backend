package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v9"
)

type OrderOfferSnapshot struct {
	OrderID                uuid.UUID                     `db:"order_id"`
	UserAddress            OrderOfferSnapshotUserAddress `db:"user_address"`
	ServiceName            string                        `db:"service_name"`
	ServiceDeliveryMethods DeliveryMethods               `db:"service_delivery_methods"`
	ServiceRules           ServiceRules                  `db:"service_rules"`
	ServiceDescription     string                        `db:"service_description"`
}

type OrderOfferSnapshotUserAddress struct {
	Coordinates null.String `json:"coordinates"`
	Province    string      `json:"province"`
	City        string      `json:"city"`
	Detail      string      `json:"detail"`
}

func (t OrderOfferSnapshotUserAddress) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *OrderOfferSnapshotUserAddress) Scan(value interface{}) error {
	if value == nil {
		*t = OrderOfferSnapshotUserAddress{}
		return nil
	}

	data, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan OrderOfferSnapshotUserAddress: value is not []byte")
	}

	return json.Unmarshal(data, t)
}
