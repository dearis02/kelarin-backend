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
			for _, origin := range cfg.Server.CORS.AllowedOrigins {
				if origin == r.Header.Get("Origin") {
					return true
				}
			}

			return false
		},
	}
}

func NewWsHub() *types.WsHub {
	return &types.WsHub{
		Clients: make(map[string]*types.WsClient),
	}
}
