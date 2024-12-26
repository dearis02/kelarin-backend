//go:build wireinject
// +build wireinject

package main

import (
	"kelarin/internal/config"
	"kelarin/internal/provider"

	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func newServer(db *sqlx.DB, config *config.Config, redis *redis.Client) (*provider.Server, error) {
	wire.Build(
		provider.RepositorySet,
		provider.ServiceSet,
		provider.HandlerSet,
		provider.NewServer,
	)

	return &provider.Server{}, nil
}
