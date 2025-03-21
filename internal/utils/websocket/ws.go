package websocket

import (
	"kelarin/internal/config"
	"kelarin/internal/types"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func NewWsUpgrader(cfg *config.Config) *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout: 5 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
			// return slices.Contains(cfg.Server.CORS.AllowedOrigins, r.Header.Get("Origin"))
		},
	}
}

func NewWsHub() *types.WsHub {
	return &types.WsHub{
		Clients: make(map[string]*types.WsClient),
	}
}
