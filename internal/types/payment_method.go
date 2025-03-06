package types

import "github.com/google/uuid"

// region repo types

type PaymentMethod struct {
	ID           uuid.UUID                 `db:"id"`
	Name         string                    `db:"name"`
	Type         PaymentMethodType         `db:"type"`
	Code         string                    `db:"code"`
	AdminFee     int32                     `db:"admin_fee"`
	AdminFeeUnit PaymentMethodAdminFeeUnit `db:"admin_fee_unit"`
	Logo         string                    `db:"logo"`
	Enabled      bool                      `db:"enabled"`
}

type PaymentMethodType string

const (
	PaymentMethodTypeVA   PaymentMethodType = "va"
	PaymentMethodTypeQRIS PaymentMethodType = "qris"
)

type PaymentMethodAdminFeeUnit string

const (
	PaymentMethodAdminFeeUnitFixed      PaymentMethodAdminFeeUnit = "fixed"
	PaymentMethodAdminFeeUnitPercentage PaymentMethodAdminFeeUnit = "percentage"
)

// endregion repo types
