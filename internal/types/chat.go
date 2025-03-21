package types

import (
	"time"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// region service types

type ChatChatRoomCreateReq struct {
	AuthUser    AuthUser      `middleware:"user"`
	ServiceID   uuid.NullUUID `json:"service_id"`
	SenderID    uuid.UUID     `json:"sender_id"`
	RecipientID uuid.UUID     `json:"recipient_id"`
	OfferID     uuid.NullUUID `json:"offer_id"`
	Tx          *sqlx.Tx
}

func (r ChatChatRoomCreateReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.RecipientID, validation.Required),
	)
}

type ChatChatRoomCreateRes struct {
	RoomID uuid.UUID `json:"room_id"`
}

type ChatSendMessageReq struct {
	RoomID            uuid.NullUUID          `json:"room_id"`
	SenderUserID      uuid.UUID              `json:"-"`
	ServiceProviderID uuid.NullUUID          `json:"service_provider_id"`
	Content           string                 `json:"content"`
	ContentType       ChatMessageContentType `json:"content_type"`
}

func (r ChatSendMessageReq) Validate() error {
	if !r.RoomID.Valid && !r.ServiceProviderID.Valid {
		return validation.NewError("room_id_or_recipient_user_id_required", "room_id or recipient_user_id is required")
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.Content, validation.Required),
		validation.Field(&r.ContentType, validation.Required),
		validation.Field(&r.ServiceProviderID, validation.Required.When(!r.RoomID.Valid)),
	)
}

type ChatIncomingMessageRes struct {
	RoomID      uuid.UUID              `json:"room_id"`
	MessageID   uuid.UUID              `json:"message_id"`
	Content     string                 `json:"content"`
	ContentType ChatMessageContentType `json:"content_type"`
	CreatedAt   time.Time              `json:"created_at"`
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

type ChatSaveSentMessageRes struct {
	RoomID    uuid.UUID `json:"room_id"`
	MessageID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatGetAllReq struct {
	AuthUser AuthUser `middleware:"user"`
}

func (r ChatGetAllReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return nil
}

type ChatContext string

const (
	ChatContextCommon  ChatContext = "common"
	ChatContextService ChatContext = "service"
	ChatContextOrder   ChatContext = "order"
)

type ChatConsumerGetAllRes struct {
	Context         ChatContext                          `json:"context"`
	RoomID          uuid.UUID                            `json:"room_id"`
	ServiceProvider ChatConsumerGetAllResServiceProvider `json:"service_provider"`
	UnreadMessages  []ChatGetAllResUnreadMessage         `json:"unread_messages"`
}

type ChatConsumerGetAllResServiceProvider struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	LogoURL string    `json:"logo_url"`
}

type ChatGetAllResUnreadMessage struct {
	ID          uuid.UUID              `json:"id"`
	Content     string                 `json:"content"`
	ContentType ChatMessageContentType `json:"content_type"`
	CreatedAt   time.Time              `json:"created_at"`
}

type ChatGetByRoomIDReq struct {
	AuthUser AuthUser  `middleware:"user"`
	TimeZone string    `header:"Time-Zone"`
	RoomID   uuid.UUID `param:"room_id"`
}

func (r ChatGetByRoomIDReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	if r.RoomID == uuid.Nil {
		return ErrIDRouteParamRequired
	}

	return nil
}

type ChatConsumerGetByRoomIDRes struct {
	Context         ChatContext                       `json:"context"`
	RoomID          uuid.UUID                         `json:"room_id"`
	OfferID         uuid.NullUUID                     `json:"offer_id"`
	Service         *ChatGetByRoomIDResService        `json:"service"`
	Order           *ChatGetByRoomIDResOrder          `json:"order"`
	ServiceProvider ChatGetByRoomIDResServiceProvider `json:"service_provider"`
	Messages        []ChatGetByRoomIDResMessage       `json:"messages"`
}

type ChatGetByRoomIDResService struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ChatGetByRoomIDResServiceProvider struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	LogoURL string    `json:"logo_url"`
}

type ChatGetByRoomIDResOrder struct {
	ID          uuid.UUID   `json:"id"`
	Status      OrderStatus `json:"status"`
	ServiceDate string      `json:"service_date"`
	ServiceTime string      `json:"service_time"`
}

type ChatGetByRoomIDResMessage struct {
	ID          uuid.UUID              `json:"id"`
	SenderID    uuid.UUID              `json:"sender_id"`
	Content     string                 `json:"content"`
	ContentType ChatMessageContentType `json:"content_type"`
	Read        bool                   `json:"read"`
	CreatedAt   time.Time              `json:"created_at"`
}

// end of region service types
