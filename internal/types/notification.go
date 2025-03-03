package types

import (
	"fmt"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

// region repo types

func FCMTokenKey(userID uuid.UUID) string {
	return fmt.Sprintf("fcm-token:%s", userID)
}

// endregion repo types

// region service types

type NotificationSendReq struct {
	Title    string
	Message  string
	IconURL  string
	BadgeURL string
	ImageURL string
	Token    string
}

type NotificationSaveTokenReq struct {
	AuthUser AuthUser `middleware:"user"`
	Token    string   `json:"token"`
}

func (r NotificationSaveTokenReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return validation.ValidateStruct(&r, validation.Field(&r.Token, validation.Required))
}

// endregion service types
