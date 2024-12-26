//go:build wireinject
// +build wireinject

package main

import (
	"kelarin/internal/config"
	"kelarin/internal/provider"

	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
)

func newServer(db *sqlx.DB, config *config.Config) (*provider.Server, error) {
	wire.Build(
		provider.RepositorySet,
		provider.ServiceSet,
		provider.HandlerSet,
		provider.NewServer,
	)

	return &provider.Server{}, nil
}
