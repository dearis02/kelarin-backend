package provider

import (
	"kelarin/internal/service"

	"github.com/google/wire"
)

var ServiceSet = wire.NewSet(service.NewUser, service.NewAuth, service.NewFile, service.NewGeocoding, service.NewServiceProvider)
