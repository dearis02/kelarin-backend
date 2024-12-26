package dbUtil

import (
	"context"
	"kelarin/internal/config"

	"github.com/go-errors/errors"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.Redis.ConString)
	if err != nil {
		return nil, errors.New(err)
	}

	client := redis.NewClient(opts)

	err = client.Ping(context.Background()).Err()
	if err != nil {
		return client, errors.New(err)
	}

	return client, nil
}
