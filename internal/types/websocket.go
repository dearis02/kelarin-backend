package types

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	sync.Mutex
	Ctx      context.Context
	AuthUser AuthUser `middleware:"user"`
	Con      *websocket.Conn
}

type WsHub struct {
	Clients map[string]*WsClient
}
