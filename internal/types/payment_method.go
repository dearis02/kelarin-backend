package types

import "github.com/google/uuid"

// region repo types

type PaymentMethod struct {
	ID           uuid.UUID                 `db:"id"`
	Name         string                    `db:"name"`
	Type         PaymentMethodType         `db:"type"`
	Code         string                    `db:"code"`
	AdminFee     float32                   `db:"admin_fee"`
	AdminFeeUnit PaymentMethodAdminFeeUnit `db:"admin_fee_unit"`
	Logo         string                    `db:"logo"`
	Enabled      bool                      `db:"enabled"`
}

type PaymentMethodType string

const (
	PaymentMethodTypeVA PaymentMethodType = "va"
	PaymentMethodTypeQR PaymentMethodType = "qr"
)

type PaymentMethodAdminFeeUnit string

const (
	PaymentMethodAdminFeeUnitFixed      PaymentMethodAdminFeeUnit = "fixed"
	PaymentMethodAdminFeeUnitPercentage PaymentMethodAdminFeeUnit = "percent"
)

// endregion repo types

// region service types

type PaymentMethodGetAllRes struct {
	ID           uuid.UUID                 `json:"id"`
	Name         string                    `json:"name"`
	Type         PaymentMethodType         `json:"type"`
	AdminFee     float32                   `json:"admin_fee"`
	AdminFeeUnit PaymentMethodAdminFeeUnit `json:"admin_fee_unit"`
	LogoURL      string                    `json:"logo_url"`
	Enabled      bool                      `json:"enabled"`
}

// endregion service types
