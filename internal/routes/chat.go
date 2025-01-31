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

func NewChat(g *gin.Engine, chatHandler handler.Chat) Chat {
	return Chat{
		g:           g,
		chatHandler: chatHandler,
	}
}

func (r *Chat) Register(m middleware.Auth) {
	r.g.GET("/v1/web-socket/chat", m.Authenticated, r.chatHandler.HandleInboundMessage)
}
