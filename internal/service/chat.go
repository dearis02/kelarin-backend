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
	"github.com/samber/lo"
)

type Chat interface {
	HandleInboundMessage(client *types.WsClient)
	CreateChatRoom(ctx context.Context, req types.ChatChatRoomCreateReq) (types.ChatChatRoomCreateRes, error)

	ConsumerGetAll(ctx context.Context, req types.ChatGetAllReq) ([]types.ChatConsumerGetAllRes, error)
	ConsumerGetByRoomID(ctx context.Context, req types.ChatGetByRoomIDReq) (types.ChatConsumerGetByRoomIDRes, error)

	ProviderGetAll(ctx context.Context, req types.ChatGetAllReq) ([]types.ChatProviderGetAllRes, error)
	ProviderGetByRoomID(ctx context.Context, req types.ChatGetByRoomIDReq) (types.ChatProviderGetByRoomIDRes, error)
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
	fileSvc             File
	orderRepo           repository.Order
	utilSvc             Util
}

func NewChat(
	db *sqlx.DB,
	serviceRepo repository.Service,
	userRepo repository.User,
	chatRoomRepo repository.ChatRoom,
	chatRoomUserRepo repository.ChatRoomUser,
	chatMessageRepo repository.ChatMessage,
	hub *types.WsHub,
	offerRepo repository.Offer,
	serviceProviderRepo repository.ServiceProvider,
	fileSvc File,
	orderRepo repository.Order,
	utilSvc Util,
) Chat {
	return &chatImpl{
		db:                  db,
		serviceRepo:         serviceRepo,
		userRepo:            userRepo,
		chatRoomRepo:        chatRoomRepo,
		chatRoomUserRepo:    chatRoomUserRepo,
		chatMessageRepo:     chatMessageRepo,
		hub:                 hub,
		offerRepo:           offerRepo,
		serviceProviderRepo: serviceProviderRepo,
		fileSvc:             fileSvc,
		orderRepo:           orderRepo,
		utilSvc:             utilSvc,
	}
}

func (s *chatImpl) HandleInboundMessage(client *types.WsClient) {
	for {
		client.Lock()

		wsRes := types.WsResponse{
			Success: false,
			Type:    types.WsResponseTypeServer,
			Code:    types.WsResponseCodeInternalServerError,
			Message: "internal server error",
		}

		m := types.ChatSendMessageReq{
			SenderUserID: client.AuthUser.ID,
		}

		_, msg, err := client.Con.ReadMessage()
		if err != nil {
			log.Error().Stack().Err(err).Send()
			wsRes.Message = "error reading message"

			res, err := wsRes.Parse()
			if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				client.Unlock()
				return
			}

			client.Con.WriteMessage(websocket.BinaryMessage, res)
			client.Con.Close()
			client.Unlock()

			return
		}

		client.Unlock()

		if err := json.Unmarshal(msg, &m); err != nil {
			log.Error().Stack().Err(err).Send()

			wsRes.Code = types.WsResponseCodeClientError
			wsRes.Message = "error parsing message"

			res, err := wsRes.Parse()
			if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			client.Con.WriteMessage(websocket.BinaryMessage, res)
			continue
		}

		if err := m.Validate(); err != nil {
			wsRes.Code = types.WsResponseCodeClientError
			wsRes.Message = "error validating message"
			wsRes.Errors = err

			res, err := wsRes.Parse()
			if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			client.Con.WriteMessage(websocket.BinaryMessage, res)
			continue
		}

		wsRes.Metadata = types.ChatSendMessageResMetadata{
			ID: m.ID,
		}

		recipientUserID := m.ServiceProviderID

		if m.RoomID.Valid {
			chatRoomUser, err := s.chatRoomUserRepo.FindByChatRoomIDAndUserID(client.Ctx, m.RoomID.UUID, client.AuthUser.ID)
			if errors.Is(err, types.ErrNoData) {
				wsRes.Code = types.WsResponseCodeChatRoomNotFound
				wsRes.Message = "chat room not found"

				res, err := wsRes.Parse()
				if err != nil {
					log.Error().Stack().Err(err).Send()
					client.Con.Close()
					return
				}

				client.Con.WriteMessage(websocket.BinaryMessage, res)
				continue
			} else if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			recipients, err := s.chatRoomUserRepo.FindRecipientByChatRoomIDs(client.Ctx, client.AuthUser.ID, uuid.UUIDs{chatRoomUser.ChatRoomID})
			if err != nil {
				log.Error().Stack().Err(err).Send()

				res, err := wsRes.Parse()
				if err != nil {
					log.Error().Stack().Err(err).Send()
					client.Con.Close()
					return
				}

				client.Con.WriteMessage(websocket.BinaryMessage, res)
				return
			}

			if len(recipients) == 0 {
				wsRes.Code = types.WsResponseCodeInternalServerError
				wsRes.Message = "recipient not found"

				res, err := wsRes.Parse()
				if err != nil {
					log.Error().Stack().Err(err).Send()
					client.Con.Close()
					return
				}

				client.Con.WriteMessage(websocket.BinaryMessage, res)
				continue
			}

			recipientUserID = uuid.NullUUID{UUID: recipients[0].UserID, Valid: true}
		} else if m.ServiceProviderID.Valid {
			serviceProvider, err := s.serviceProviderRepo.FindByID(client.Ctx, m.ServiceProviderID.UUID)
			if errors.Is(err, types.ErrNoData) {
				wsRes.Code = types.WsResponseCodeChatRecipientNotFound
				wsRes.Message = "service provider not found"

				res, err := wsRes.Parse()
				if err != nil {
					log.Error().Stack().Err(err).Send()
					client.Con.Close()
					return
				}

				client.Con.WriteMessage(websocket.BinaryMessage, res)
				continue
			} else if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			recipientUserID = uuid.NullUUID{UUID: serviceProvider.UserID, Valid: true}
		}

		if recipientUserID.UUID == client.AuthUser.ID {
			wsRes.Code = types.WsResponseCodeClientError
			wsRes.Message = "cannot send message to yourself"

			res, err := wsRes.Parse()
			if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			client.Con.WriteMessage(websocket.BinaryMessage, res)
			continue
		}

		req := types.ChatSaveSentMessageReq{
			AuthUser:        client.AuthUser,
			RoomID:          m.RoomID,
			RecipientUserID: recipientUserID,
			Content:         m.Content,
			ContentType:     m.ContentType,
		}

		saveSentMsgRes, err := s.SaveSentMessage(client.Ctx, req)
		if err != nil {
			log.Error().Stack().Err(err).Send()

			res, err := wsRes.Parse()
			if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			client.Con.WriteMessage(websocket.BinaryMessage, res)
			continue
		}

		targetClient, ok := s.hub.Clients[recipientUserID.UUID.String()]
		if !ok {
			wsRes.Success = true
			wsRes.Code = types.WsResponseCodeChatRecipientOffline
			wsRes.Message = "recipient is offline"

			res, err := wsRes.Parse()
			if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			client.Con.WriteMessage(websocket.BinaryMessage, res)
			continue
		}

		if targetClient.Con == nil {
			log.Error().Stack().Err(errors.New("target client connection is nil")).Send()

			res, err := wsRes.Parse()
			if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			client.Con.WriteMessage(websocket.BinaryMessage, res)
			continue
		}

		// ping target client
		if err := targetClient.Con.WriteMessage(websocket.PingMessage, nil); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) || websocket.IsCloseError(err, websocket.CloseGoingAway) {
				wsRes.Success = true
				wsRes.Message = "recipient is offline"
				wsRes.Code = types.WsResponseCodeChatRecipientOffline
			} else {
				log.Error().Stack().Err(err).Send()
			}

			res, err := wsRes.Parse()
			if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			client.Con.WriteMessage(websocket.BinaryMessage, res)
			continue
		}

		incomingMsg := types.ChatIncomingMessageRes{
			RoomID:      saveSentMsgRes.RoomID,
			MessageID:   saveSentMsgRes.MessageID,
			Content:     m.Content,
			ContentType: m.ContentType,
			CreatedAt:   saveSentMsgRes.CreatedAt,
		}

		wsIncomingMsgRes := types.WsResponse{
			Success: true,
			Type:    types.WsResponseTypeChatIncomingMessage,
			Code:    types.WsResponseCodeSuccess,
			Message: "success",
			Data:    incomingMsg,
		}

		res, err := wsIncomingMsgRes.Parse()
		if err != nil {
			log.Error().Stack().Err(err).Send()
			client.Con.Close()
			return
		}

		if err := targetClient.Con.WriteMessage(websocket.BinaryMessage, res); err != nil {
			log.Error().Stack().Err(err).Send()

			errRes, err := wsRes.Parse()
			if err != nil {
				log.Error().Stack().Err(err).Send()
				client.Con.Close()
				return
			}

			client.Con.WriteMessage(websocket.BinaryMessage, errRes)
			continue
		}

		wsRes = types.WsResponse{
			Success: true,
			Type:    types.WsResponseTypeServer,
			Code:    types.WsResponseCodeSuccess,
			Message: "success",
			Metadata: types.ChatSendMessageResMetadata{
				ID: m.ID,
			},
		}

		res, err = wsRes.Parse()
		if err != nil {
			log.Error().Stack().Err(err).Send()
			client.Con.Close()
			return
		}

		client.Con.WriteMessage(websocket.BinaryMessage, res)
	}
}

func (s *chatImpl) SaveSentMessage(ctx context.Context, req types.ChatSaveSentMessageReq) (types.ChatSaveSentMessageRes, error) {
	res := types.ChatSaveSentMessageRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	userSender, err := s.userRepo.FindByID(ctx, req.AuthUser.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(fmt.Sprintf("user not found: id %d", req.AuthUser.ID))
	} else if err != nil {
		return res, err
	}

	tx, err := dbUtil.NewSqlxTx(ctx, s.db, nil)
	if err != nil {
		return res, err
	}

	defer tx.Rollback()

	var chatRoomID uuid.UUID
	if req.RoomID.Valid {
		chatRoom, err := s.chatRoomRepo.FindByID(ctx, req.RoomID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.New(fmt.Sprintf("chat room not found: id %s", req.RoomID.UUID))
		} else if err != nil {
			return res, err
		}

		chatRoomID = chatRoom.ID
	} else {
		recipient, err := s.userRepo.FindByID(ctx, req.RecipientUserID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.New(fmt.Sprintf("recipient not found: id %s", req.RecipientUserID.UUID))
		} else if err != nil {
			return res, err
		}

		newChatRoom := types.ChatChatRoomCreateReq{
			AuthUser:    req.AuthUser,
			SenderID:    userSender.ID,
			RecipientID: recipient.ID,
			Tx:          tx,
		}

		newChatRoomRes, err := s.CreateChatRoom(ctx, newChatRoom)
		if err != nil {
			return res, err
		}

		chatRoomID = newChatRoomRes.RoomID
	}

	timeNow := time.Now()

	chatMessageID, err := uuid.NewV7()
	if err != nil {
		return res, errors.New(err)
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
		return res, err
	}

	if err := tx.Commit(); err != nil {
		return res, errors.New(err)
	}

	res = types.ChatSaveSentMessageRes{
		RoomID:    chatRoomID,
		MessageID: chatMessageID,
		CreatedAt: chatMessage.CreatedAt,
	}

	return res, nil
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

	res.RoomID = chatRoomID

	return res, nil
}

func (s *chatImpl) ConsumerGetAll(ctx context.Context, req types.ChatGetAllReq) ([]types.ChatConsumerGetAllRes, error) {
	res := []types.ChatConsumerGetAllRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	chatRoomUsers, err := s.chatRoomUserRepo.FindByUserID(ctx, req.AuthUser.ID)
	if err != nil {
		return res, err
	}

	chatRoomIDs := uuid.UUIDs{}
	for _, chatRoomUser := range chatRoomUsers {
		chatRoomIDs = append(chatRoomIDs, chatRoomUser.ChatRoomID)
	}

	recipients, err := s.chatRoomUserRepo.FindRecipientByChatRoomIDs(ctx, req.AuthUser.ID, chatRoomIDs)
	if err != nil {
		return res, err
	}

	userIDs := lo.Map(recipients, func(recipient types.ChatRoomUser, _ int) uuid.UUID {
		return recipient.UserID
	})

	userIDs = lo.Uniq(userIDs)

	serviceProviders, err := s.serviceProviderRepo.FindByUserIDs(ctx, userIDs)
	if err != nil {
		return res, err
	}

	unreadMsgs, err := s.chatMessageRepo.CountUnreadReceivedByChatRoomIDs(ctx, req.AuthUser.ID, chatRoomIDs)
	if err != nil {
		return res, err
	}

	latestMessages, err := s.chatMessageRepo.FindLatestByChatRoomIDs(ctx, chatRoomIDs)
	if err != nil {
		return res, err
	}

	reqTz, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	for _, room := range chatRoomUsers {
		recipientChatRoom, exs := lo.Find(recipients, func(recipient types.ChatRoomUser) bool {
			return recipient.ChatRoomID == room.ChatRoomID
		})
		if !exs {
			continue
		}

		provider, exs := lo.Find(serviceProviders, func(provider types.ServiceProvider) bool {
			return provider.UserID == recipientChatRoom.UserID
		})
		if !exs {
			continue
		}

		logoURL, err := s.fileSvc.GetS3PresignedURL(ctx, provider.LogoImage)
		if err != nil {
			return res, err
		}

		chatCtx := types.ChatContextCommon
		if room.OfferID.Valid {
			chatCtx = types.ChatContextOrder
		} else if room.ServiceID.Valid {
			chatCtx = types.ChatContextService
		}

		unreadMsgCount := 0
		unreadMsg, exs := lo.Find(unreadMsgs, func(unreadMsg types.ChatMessageCountUnread) bool {
			return unreadMsg.ChatRoomID == room.ChatRoomID
		})

		if exs {
			unreadMsgCount = unreadMsg.Count
		}

		msg, exs := lo.Find(latestMessages, func(msg types.ChatMessage) bool {
			return msg.ChatRoomID == room.ChatRoomID
		})

		if exs {
			res = append(res, types.ChatConsumerGetAllRes{
				Context: chatCtx,
				RoomID:  room.ChatRoomID,
				ServiceProvider: types.ChatConsumerGetAllResServiceProvider{
					ID:      provider.ID,
					Name:    provider.Name,
					LogoURL: logoURL,
				},
				UnreadMessageCount: unreadMsgCount,
				LatestMessage: &types.ChatGetAllResLatestMessage{
					ID:          msg.ID,
					Content:     msg.Content,
					ContentType: msg.ContentType,
					Read:        msg.Read,
					CreatedAt:   msg.CreatedAt.In(reqTz),
				},
			})
		} else {
			res = append(res, types.ChatConsumerGetAllRes{
				Context: chatCtx,
				RoomID:  room.ChatRoomID,
				ServiceProvider: types.ChatConsumerGetAllResServiceProvider{
					ID:      provider.ID,
					Name:    provider.Name,
					LogoURL: logoURL,
				},
				UnreadMessageCount: unreadMsgCount,
			})
		}
	}

	return res, nil
}

func (s *chatImpl) ConsumerGetByRoomID(ctx context.Context, req types.ChatGetByRoomIDReq) (types.ChatConsumerGetByRoomIDRes, error) {
	res := types.ChatConsumerGetByRoomIDRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	chatRoom, err := s.chatRoomRepo.FindByID(ctx, req.RoomID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "chat room not found"})
	} else if err != nil {
		return res, err
	}

	messages, err := s.chatMessageRepo.FindByChatRoomID(ctx, chatRoom.ID)
	if err != nil {
		return res, err
	}

	recipient, err := s.chatRoomUserRepo.FindRecipientByChatRoomID(ctx, req.AuthUser.ID, chatRoom.ID)
	if err != nil {
		return res, err
	}

	provider, err := s.serviceProviderRepo.FindByUserID(ctx, recipient.UserID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(fmt.Sprintf("service provider not found: user_id %s", recipient.UserID))
	} else if err != nil {
		return res, err
	}

	logoURL, err := s.fileSvc.GetS3PresignedURL(ctx, provider.LogoImage)
	if err != nil {
		return res, err
	}

	res = types.ChatConsumerGetByRoomIDRes{
		Context: types.ChatContextCommon,
		RoomID:  chatRoom.ID,
		ServiceProvider: types.ChatGetByRoomIDResServiceProvider{
			ID:      provider.ID,
			Name:    provider.Name,
			LogoURL: logoURL,
		},
		Messages: []types.ChatGetByRoomIDResMessage{},
	}

	reqTz, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	if chatRoom.OfferID.Valid {
		order, err := s.orderRepo.FindByOfferID(ctx, chatRoom.OfferID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.Errorf("order not found: offer_id %s", chatRoom.OfferID.UUID)
		} else if err != nil {
			return res, err
		}

		res.Context = types.ChatContextOrder
		res.OfferID = uuid.NullUUID{UUID: order.OfferID, Valid: true}
		res.Order = &types.ChatGetByRoomIDResOrder{
			ID:          order.ID,
			Status:      order.Status,
			ServiceDate: order.ServiceDate.Format(time.DateOnly),
			ServiceTime: order.ServiceTime.In(reqTz).Format(time.TimeOnly),
		}

		res.Service = &types.ChatGetByRoomIDResService{
			ID:   order.ServiceID,
			Name: order.ServiceName,
		}
	} else if chatRoom.ServiceID.Valid {
		service, err := s.serviceRepo.FindByID(ctx, chatRoom.ServiceID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.Errorf("service not found: id %s", chatRoom.ServiceID.UUID)
		} else if err != nil {
			return res, err
		}

		res.Context = types.ChatContextService
		res.Service = &types.ChatGetByRoomIDResService{
			ID:   service.ID,
			Name: service.Name,
		}
	}

	for _, message := range messages {
		res.Messages = append(res.Messages, types.ChatGetByRoomIDResMessage{
			ID:          message.ID,
			IsSender:    message.UserID == req.AuthUser.ID,
			Content:     message.Content,
			ContentType: message.ContentType,
			Read:        message.Read,
			CreatedAt:   message.CreatedAt.In(reqTz),
		})
	}

	return res, nil
}

func (s *chatImpl) ProviderGetAll(ctx context.Context, req types.ChatGetAllReq) ([]types.ChatProviderGetAllRes, error) {
	res := []types.ChatProviderGetAllRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	chatRoomUsers, err := s.chatRoomUserRepo.FindByUserID(ctx, req.AuthUser.ID)
	if err != nil {
		return res, err
	}

	chatRoomIDs := uuid.UUIDs{}
	for _, chatRoomUser := range chatRoomUsers {
		chatRoomIDs = append(chatRoomIDs, chatRoomUser.ChatRoomID)
	}

	recipients, err := s.chatRoomUserRepo.FindRecipientByChatRoomIDs(ctx, req.AuthUser.ID, chatRoomIDs)
	if err != nil {
		return res, err
	}

	userIDs := lo.Map(recipients, func(recipient types.ChatRoomUser, _ int) uuid.UUID {
		return recipient.UserID
	})

	userIDs = lo.Uniq(userIDs)

	consumers, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return res, err
	}

	unreadMsgs, err := s.chatMessageRepo.CountUnreadReceivedByChatRoomIDs(ctx, req.AuthUser.ID, chatRoomIDs)
	if err != nil {
		return res, err
	}

	latestMessages, err := s.chatMessageRepo.FindLatestByChatRoomIDs(ctx, chatRoomIDs)
	if err != nil {
		return res, err
	}

	reqTz, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	for _, room := range chatRoomUsers {
		recipientChatRoom, exs := lo.Find(recipients, func(recipient types.ChatRoomUser) bool {
			return recipient.ChatRoomID == room.ChatRoomID
		})
		if !exs {
			continue
		}

		consumer, exs := lo.Find(consumers, func(consumer types.User) bool {
			return consumer.ID == recipientChatRoom.UserID
		})
		if !exs {
			continue
		}

		chatCtx := types.ChatContextCommon
		if room.OfferID.Valid {
			chatCtx = types.ChatContextOrder
		} else if room.ServiceID.Valid {
			chatCtx = types.ChatContextService
		}

		unreadMsgCount := 0
		unreadMsg, exs := lo.Find(unreadMsgs, func(unreadMsg types.ChatMessageCountUnread) bool {
			return unreadMsg.ChatRoomID == room.ChatRoomID
		})

		if exs {
			unreadMsgCount = unreadMsg.Count
		}

		msg, exs := lo.Find(latestMessages, func(msg types.ChatMessage) bool {
			return msg.ChatRoomID == room.ChatRoomID
		})

		if exs {
			res = append(res, types.ChatProviderGetAllRes{
				Context:            chatCtx,
				RoomID:             room.ChatRoomID,
				UnreadMessageCount: unreadMsgCount,
				Consumer: types.ChatProviderGetAllResConsumer{
					ID:   consumer.ID,
					Name: consumer.Name,
				},
				LatestMessage: &types.ChatGetAllResLatestMessage{
					ID:          msg.ID,
					Content:     msg.Content,
					ContentType: msg.ContentType,
					Read:        msg.Read,
					CreatedAt:   msg.CreatedAt.In(reqTz),
				},
			})
		} else {
			res = append(res, types.ChatProviderGetAllRes{
				Context:            chatCtx,
				RoomID:             room.ChatRoomID,
				UnreadMessageCount: unreadMsgCount,
				Consumer: types.ChatProviderGetAllResConsumer{
					ID:   consumer.ID,
					Name: consumer.Name,
				},
			})
		}
	}

	return res, nil
}

func (s *chatImpl) ProviderGetByRoomID(ctx context.Context, req types.ChatGetByRoomIDReq) (types.ChatProviderGetByRoomIDRes, error) {
	res := types.ChatProviderGetByRoomIDRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	chatRoom, err := s.chatRoomRepo.FindByID(ctx, req.RoomID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusNotFound, Message: "chat room not found"})
	} else if err != nil {
		return res, err
	}

	messages, err := s.chatMessageRepo.FindByChatRoomID(ctx, chatRoom.ID)
	if err != nil {
		return res, err
	}

	recipient, err := s.chatRoomUserRepo.FindRecipientByChatRoomID(ctx, req.AuthUser.ID, chatRoom.ID)
	if err != nil {
		return res, err
	}

	consumer, err := s.userRepo.FindByID(ctx, recipient.UserID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(fmt.Sprintf("service provider not found: user_id %s", recipient.UserID))
	} else if err != nil {
		return res, err
	}

	res = types.ChatProviderGetByRoomIDRes{
		Context: types.ChatContextCommon,
		RoomID:  chatRoom.ID,
		Consumer: types.ChatProviderGetByRoomIDResConsumer{
			ID:   consumer.ID,
			Name: consumer.Name,
		},
		Messages: []types.ChatGetByRoomIDResMessage{},
	}

	reqTz, err := s.utilSvc.ParseUserTimeZone(req.TimeZone)
	if err != nil {
		return res, err
	}

	if chatRoom.OfferID.Valid {
		order, err := s.orderRepo.FindByOfferID(ctx, chatRoom.OfferID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.Errorf("order not found: offer_id %s", chatRoom.OfferID.UUID)
		} else if err != nil {
			return res, err
		}

		res.Context = types.ChatContextOrder
		res.OfferID = uuid.NullUUID{UUID: order.OfferID, Valid: true}
		res.Order = &types.ChatGetByRoomIDResOrder{
			ID:          order.ID,
			Status:      order.Status,
			ServiceDate: order.ServiceDate.Format(time.DateOnly),
			ServiceTime: order.ServiceTime.In(reqTz).Format(time.TimeOnly),
		}

		res.Service = &types.ChatGetByRoomIDResService{
			ID:   order.ServiceID,
			Name: order.ServiceName,
		}
	} else if chatRoom.ServiceID.Valid {
		service, err := s.serviceRepo.FindByID(ctx, chatRoom.ServiceID.UUID)
		if errors.Is(err, types.ErrNoData) {
			return res, errors.Errorf("service not found: id %s", chatRoom.ServiceID.UUID)
		} else if err != nil {
			return res, err
		}

		res.Context = types.ChatContextService
		res.Service = &types.ChatGetByRoomIDResService{
			ID:   service.ID,
			Name: service.Name,
		}
	}

	for _, message := range messages {
		res.Messages = append(res.Messages, types.ChatGetByRoomIDResMessage{
			ID:          message.ID,
			IsSender:    message.UserID == req.AuthUser.ID,
			Content:     message.Content,
			ContentType: message.ContentType,
			Read:        message.Read,
			CreatedAt:   message.CreatedAt.In(reqTz),
		})
	}

	return res, nil
}
