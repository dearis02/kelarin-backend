package provider

import (
	"kelarin/internal/handler"

	"github.com/google/wire"
)

var HandlerSet = wire.NewSet(handler.NewUser, handler.NewAuth)
