package repository

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/redis/go-redis/v9"
)

type Session interface {
	Set(ctx context.Context, key string, val string, duration time.Duration) error
}

type sessionImpl struct {
	redis *redis.Client
}

func NewSession(redis *redis.Client) Session {
	return &sessionImpl{redis: redis}
}

func (s *sessionImpl) Set(ctx context.Context, key string, val string, duration time.Duration) error {
	if err := s.redis.Set(ctx, key, val, duration).Err(); err != nil {
		return errors.New(err)
	}

	return nil
}
