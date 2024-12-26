package provider

import (
	"kelarin/internal/repository"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(repository.NewUser)
