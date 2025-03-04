package service

import (
	"context"
	"encoding/json"
	"fmt"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type Chat interface {
	HandleInboundMessage(client *types.WsClient)
	CreateChatRoom(ctx context.Context, req types.ChatChatRoomCreateReq) (types.ChatChatRoomCreateRes, error)
}

type chatImpl struct {
	db                  *sqlx.DB
	serviceRepo         repository.Service
	userRepo            repository.User
	chatRoomRepo        repository.ChatRoom
	chatRoomUserRepo    repository.ChatRoomUser
	chatMessageRepo     repository.ChatMessage
	hub                 *types.WsHub
	offerRepo           repository.Offer
	serviceProviderRepo repository.ServiceProvider
}

func NewChat(userRepo repository.User, chatRoomRepo repository.ChatRoom, chatRoomUserRepo repository.ChatRoomUser, chatMessageRepo repository.ChatMessage, offerRepo repository.Offer, serviceProviderRepo repository.ServiceProvider) Chat {
	return &chatImpl{
		userRepo:            userRepo,
		chatRoomRepo:        chatRoomRepo,
		chatRoomUserRepo:    chatRoomUserRepo,
		chatMessageRepo:     chatMessageRepo,
		offerRepo:           offerRepo,
		serviceProviderRepo: serviceProviderRepo,
	}
}

func (s *chatImpl) HandleInboundMessage(client *types.WsClient) {
	for {
		client.Lock()

		m := types.ChatSendMessageReq{
			FromUserID: client.AuthUser.ID,
		}
		_, msg, err := client.Con.ReadMessage()
		if err != nil {
			log.Error().Stack().Err(err).Send()
			client.Con.WriteMessage(websocket.BinaryMessage, []byte("error reading message"))
			client.Lock()
			continue
		}

		if err := json.Unmarshal(msg, &m); err != nil {
			log.Error().Stack().Err(err).Send()
			client.Con.WriteMessage(websocket.BinaryMessage, []byte("error parsing message"))
			client.Lock()
			continue
		}

		req := types.ChatSaveSentMessageReq{
			AuthUser:        client.AuthUser,
			RoomID:          m.RoomID,
			RecipientUserID: uuid.NullUUID{UUID: m.ToUserID, Valid: true},
			Content:         m.Content,
			ContentType:     m.ContentType,
		}
		if err := s.SaveSentMessage(client.Ctx, req); err != nil {
			log.Error().Stack().Err(err).Send()
			client.Con.WriteMessage(websocket.BinaryMessage, []byte("error saving message"))
			client.Unlock()
			continue
		}

		targetClient, ok := s.hub.Clients[m.ToUserID.String()]
		if !ok {
			client.Con.WriteMessage(websocket.BinaryMessage, []byte("recipient is not connected"))
			client.Unlock()
			continue
		}

		targetClient.Lock()
		if targetClient.Con == nil {
			targetClient.Unlock()

			client.Con.WriteMessage(websocket.BinaryMessage, []byte("recipient is not connected"))
			client.Unlock()
			continue
		}

		if err := targetClient.Con.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			log.Error().Stack().Err(err).Send()
			targetClient.Unlock()

			client.Con.WriteMessage(websocket.BinaryMessage, []byte("error sending message"))
			client.Unlock()
			continue
		}

		client.Unlock()
	}
}

func (s *chatImpl) SaveSentMessage(ctx context.Context, req types.ChatSaveSentMessageReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	userSender, err := s.userRepo.FindByID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return errors.New(fmt.Sprintf("user not found: id %d", req.AuthUser.ID))
	} else if err != nil {
		return err
	}

	tx, err := dbUtil.NewSqlxTx(ctx, s.db, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var chatRoomID uuid.UUID
	if req.RoomID.Valid {
		chatRoom, err := s.chatRoomRepo.FindByID(ctx, req.RoomID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return errors.New(fmt.Sprintf("chat room not found: id %s", req.RoomID.UUID))
		} else if err != nil {
			return err
		}

		chatRoomID = chatRoom.ID
	} else {
		recipient, err := s.userRepo.FindByID(ctx, req.RecipientUserID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return errors.New(fmt.Sprintf("recipient not found: id %s", req.RecipientUserID.UUID))
		} else if err != nil {
			return err
		}

		newChatRoom := types.ChatChatRoomCreateReq{
			AuthUser:    req.AuthUser,
			SenderID:    userSender.ID,
			RecipientID: recipient.ID,
			Tx:          tx,
		}

		newChatRoomRes, err := s.CreateChatRoom(ctx, newChatRoom)
		if err != nil {
			return err
		}

		chatRoomID = newChatRoomRes.RoomID
	}

	timeNow := time.Now()

	chatMessageID, err := uuid.NewV7()
	if err != nil {
		return errors.New(err)
	}

	chatMessage := types.ChatMessage{
		ID:          chatMessageID,
		ChatRoomID:  chatRoomID,
		UserID:      userSender.ID,
		Content:     req.Content,
		ContentType: req.ContentType,
		CreatedAt:   timeNow,
	}

	if err := s.chatMessageRepo.CreateTx(ctx, tx, chatMessage); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.New(err)
	}

	return nil
}

func (s *chatImpl) CreateChatRoom(ctx context.Context, req types.ChatChatRoomCreateReq) (types.ChatChatRoomCreateRes, error) {
	res := types.ChatChatRoomCreateRes{}

	err := req.Validate()
	if err != nil {
		return res, err
	}

	timeNow := time.Now()

	chatRoomID, err := uuid.NewV7()
	if err != nil {
		return res, errors.New(err)
	}

	chatRoom := types.ChatRoom{
		ID:        chatRoomID,
		CreatedAt: timeNow,
	}

	if req.ServiceID.Valid {
		service, err := s.serviceRepo.FindByID(ctx, req.ServiceID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.New(fmt.Sprintf("service not found: id %s", req.ServiceID.UUID))
		} else if err != nil {
			return res, err
		}

		chatRoom.ServiceID = uuid.NullUUID{UUID: service.ID, Valid: true}
	}
	if req.OfferID.Valid {
		var offer types.Offer
		switch req.AuthUser.Role {
		case types.UserRoleConsumer:
			offer, err = s.offerRepo.FindByIDAndUserID(ctx, req.OfferID.UUID, req.AuthUser.ID)
			if errors.Is(err, types.ErrNoData) {
				return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "offer not found"})
			} else if err != nil {
				return res, err
			}
		case types.UserRoleServiceProvider:
			provider, err := s.serviceProviderRepo.FindByUserID(ctx, req.AuthUser.ID)
			if errors.Is(err, types.ErrNoData) {
				return res, errors.Errorf("provider not found: user_id %s", req.AuthUser.ID)
			} else if err != nil {
				return res, err
			}

			offer, err = s.offerRepo.FindByIDAndServiceProviderID(ctx, req.OfferID.UUID, provider.ID)
			if errors.Is(err, types.ErrNoData) {
				return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "offer not found"})
			} else if err != nil {
				return res, err
			}
		default:
			return res, errors.New(types.AppErr{Code: http.StatusForbidden})
		}

		chatRoom.OfferID = uuid.NullUUID{UUID: offer.ID, Valid: true}
	}

	chatRoomUser := []types.ChatRoomUser{
		{
			ChatRoomID: chatRoomID,
			UserID:     req.SenderID,
			CreatedAt:  timeNow,
		},
		{
			ChatRoomID: chatRoomID,
			UserID:     req.RecipientID,
			CreatedAt:  timeNow,
		},
	}

	if err := s.chatRoomRepo.CreateTx(ctx, req.Tx, chatRoom); err != nil {
		return res, err
	}

	if err := s.chatRoomUserRepo.CreateTx(ctx, req.Tx, chatRoomUser); err != nil {
		return res, err
	}

	return res, nil
}
