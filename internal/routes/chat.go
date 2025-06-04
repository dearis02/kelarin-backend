package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Chat struct {
	g           *gin.Engine
	chatHandler handler.Chat
}

func NewChat(g *gin.Engine, chatHandler handler.Chat) *Chat {
	return &Chat{
		g:           g,
		chatHandler: chatHandler,
	}
}

func (r *Chat) Register(m middleware.Auth) {
	r.g.GET("/v1/web-socket/chat", m.WS, r.chatHandler.HandleInboundMessage)

	r.g.PUT("/consumer/v1/chat-rooms", m.Consumer, r.chatHandler.ConsumerCreateChatRoom)
	r.g.GET("/consumer/v1/chats", m.Consumer, r.chatHandler.ConsumerGetAll)
	r.g.GET("/consumer/v1/chats/:room_id", m.Consumer, r.chatHandler.ConsumerGetByRoomID)

	r.g.GET("/provider/v1/chats", m.ServiceProvider, r.chatHandler.ProviderGetAll)
	r.g.GET("/provider/v1/chats/:room_id", m.ServiceProvider, r.chatHandler.ProviderGetByRoomID)

}
