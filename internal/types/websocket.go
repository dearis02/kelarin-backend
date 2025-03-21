package types

import (
	"context"
	"encoding/json"
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

type WsResponse struct {
	Success bool           `json:"success"`
	Type    WsResponseType `json:"type"`
	Code    WsResponseCode `json:"code"`
	Message string         `json:"message"`
	Data    any            `json:"data,omitempty"`
	Errors  any            `json:"errors,omitempty"`
}

func (r WsResponse) Parse() ([]byte, error) {
	return json.Marshal(r)
}

type WsResponseType string

const (
	WsResponseTypeServer              WsResponseType = "server"
	WsResponseTypeChatIncomingMessage WsResponseType = "incoming_message"
)

type WsResponseCode int16

const (
	WsResponseCodeSuccess             WsResponseCode = 2000
	WsResponseCodeInternalServerError WsResponseCode = 2001

	WsResponseCodeClientError           WsResponseCode = 2400
	WsResponseCodeChatRoomNotFound      WsResponseCode = 2402
	WsResponseCodeChatRecipientNotFound WsResponseCode = 2403
	WsResponseCodeChatRecipientOffline  WsResponseCode = 2404
)
