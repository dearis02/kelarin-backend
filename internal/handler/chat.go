package handler

import (
	"context"
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Chat interface {
	HandleInboundMessage(c *gin.Context)

	ConsumerGetAll(c *gin.Context)
	ConsumerGetByRoomID(c *gin.Context)

	ProviderGetAll(c *gin.Context)
	ProviderGetByRoomID(c *gin.Context)
}

type chatImpl struct {
	wsUpgrader     *websocket.Upgrader
	hub            *types.WsHub
	chatService    service.Chat
	authMiddleware middleware.Auth
}

func NewChat(upgrader *websocket.Upgrader, chatService service.Chat, hub *types.WsHub, authMw middleware.Auth) Chat {
	return &chatImpl{
		wsUpgrader:     upgrader,
		hub:            hub,
		chatService:    chatService,
		authMiddleware: authMw,
	}
}

func (h *chatImpl) HandleInboundMessage(ctx *gin.Context) {
	con, err := h.wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Error().Stack().Str("message", "error upgrading connection").Err(err).Send()
		return
	}

	client := types.WsClient{
		Ctx: context.Background(),
		Con: con,
	}

	if err = h.authMiddleware.BindWithRequest(ctx, &client); err != nil {
		log.Error().Stack().Err(err).Send()
		return
	}

	h.hub.Clients[client.AuthUser.ID.String()] = &client

	go h.chatService.HandleInboundMessage(&client)
}

func (h *chatImpl) ConsumerGetAll(c *gin.Context) {
	var req types.ChatGetAllReq
	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.chatService.ConsumerGetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *chatImpl) ConsumerGetByRoomID(c *gin.Context) {
	var req types.ChatGetByRoomIDReq
	if err := req.RoomID.UnmarshalText([]byte(c.Param("room_id"))); err != nil {
		c.Error(err)
		return
	}

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.chatService.ConsumerGetByRoomID(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *chatImpl) ProviderGetAll(c *gin.Context) {
	var req types.ChatGetAllReq
	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.chatService.ProviderGetAll(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *chatImpl) ProviderGetByRoomID(c *gin.Context) {
	var req types.ChatGetByRoomIDReq
	if err := req.RoomID.UnmarshalText([]byte(c.Param("room_id"))); err != nil {
		c.Error(err)
		return
	}

	if err := h.authMiddleware.BindWithRequest(c, &req); err != nil {
		c.Error(err)
		return
	}

	res, err := h.chatService.ProviderGetByRoomID(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		StatusCode: http.StatusOK,
		Data:       res,
	})
}
