package types

import (
	"time"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// region repo types

type Payment struct {
	ID              uuid.UUID       `db:"id"`
	PaymentMethodID uuid.UUID       `db:"payment_method_id"`
	UserID          uuid.UUID       `db:"user_id"`
	Amount          decimal.Decimal `db:"amount"`
	AdminFee        int32           `db:"admin_fee"`
	PlatformFee     int32           `db:"platform_fee"`
	PaymentLink     string          `db:"payment_link"`
	Status          PaymentStatus   `db:"status"`
	CreatedAt       time.Time       `db:"created_at"`
}

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusCanceled PaymentStatus = "canceled"
	PaymentStatusExpired  PaymentStatus = "expired"
	PaymentStatusFailed   PaymentStatus = "failed"
)

// endregion repo types

// region service types

type PaymentCreateReq struct {
	AuthUser        AuthUser  `middleware:"user"`
	OrderID         uuid.UUID `json:"order_id"`
	PaymentMethodID uuid.UUID `json:"payment_method_id"`
}

func (r PaymentCreateReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.OrderID, validation.Required),
		validation.Field(&r.PaymentMethodID, validation.Required),
	)
}

type PaymentCreateRes struct {
	PaymentLink string `json:"payment_link"`
}

type PaymentMidtransNotificationReq struct {
	FraudStatus       string                    `json:"fraud_status"`
	GrossAmount       string                    `json:"gross_amount"`
	OrderID           string                    `json:"order_id"`
	PaymentType       string                    `json:"payment_type"`
	SignatureKey      string                    `json:"signature_key"`
	StatusCode        string                    `json:"status_code"`
	TransactionID     string                    `json:"transaction_id"`
	TransactionStatus MidtransTransactionStatus `json:"transaction_status"`
}

type MidtransTransactionStatus string

const (
	MidtransTransactionStatusPending    MidtransTransactionStatus = "pending"
	MidtransTransactionStatusSettlement MidtransTransactionStatus = "settlement"
	MidtransTransactionStatusCancel     MidtransTransactionStatus = "cancel"
	MidtransTransactionStatusExpire     MidtransTransactionStatus = "expire"
	MidtransTransactionStatusFailure    MidtransTransactionStatus = "failure"
	MidtransTransactionStatusDeny       MidtransTransactionStatus = "deny"
)

// endregion service types
