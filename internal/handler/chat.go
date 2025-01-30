package handler

import (
	"kelarin/internal/middleware"
	"kelarin/internal/service"
	"kelarin/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Chat interface {
	HandleInboundMessage(c *gin.Context)
}

type chatImpl struct {
	wsUpgrader     websocket.Upgrader
	hub            *types.WsHub
	chatService    service.Chat
	authMiddleware middleware.Auth
}

func NewChat(upgrader websocket.Upgrader, chatService service.Chat, hub *types.WsHub) Chat {
	return &chatImpl{
		wsUpgrader:  upgrader,
		hub:         hub,
		chatService: chatService,
	}
}

func (h *chatImpl) HandleInboundMessage(ctx *gin.Context) {
	con, err := h.wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Error().Stack().Err(err).Send()
		return
	}

	client := types.WsClient{
		Ctx: ctx.Request.Context(),
		Con: con,
	}

	if err = h.authMiddleware.BindWithRequest(ctx, &client); err != nil {
		log.Error().Stack().Err(err).Send()
		return
	}

	h.hub.Clients[client.AuthUser.ID.String()] = &client

	go h.chatService.HandleInboundMessage(&client)
}
