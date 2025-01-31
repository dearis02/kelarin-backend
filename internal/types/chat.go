package types

import (
	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// region service types

type ChatChatRoomCreateReq struct {
	ServiceID   uuid.NullUUID `json:"service_id"`
	SenderID    uuid.UUID     `json:"sender_id"`
	RecipientID uuid.UUID     `json:"recipient_id"`
	Tx          *sqlx.Tx
}

func (r ChatChatRoomCreateReq) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.RecipientID, validation.Required),
	)
}

type ChatChatRoomCreateRes struct {
	RoomID uuid.UUID `json:"room_id"`
}

type ChatSendMessageReq struct {
	RoomID      uuid.NullUUID          `json:"room_id"`
	FromUserID  uuid.UUID              `json:"-"`
	ToUserID    uuid.UUID              `json:"to_user_id"`
	Content     string                 `json:"content"`
	ContentType ChatMessageContentType `json:"content_type"`
}

type ChatSaveSentMessageReq struct {
	AuthUser        AuthUser               `middleware:"user"`
	RoomID          uuid.NullUUID          `json:"room_id"`
	RecipientUserID uuid.NullUUID          `json:"recipient_user_id"`
	Content         string                 `json:"content"`
	ContentType     ChatMessageContentType `json:"content_type"`
}

func (r ChatSaveSentMessageReq) Validate() error {
	if r.AuthUser == (AuthUser{}) {
		return errors.New("AuthUser is required")
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.Content, validation.Required),
		validation.Field(&r.ContentType, validation.Required),
		validation.Field(&r.RecipientUserID, validation.Required.When(!r.RoomID.Valid)),
	)
}

// end of region service types
