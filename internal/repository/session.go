package repository

import (
	"context"
	"kelarin/internal/types"
	"time"

	"github.com/go-errors/errors"
	"github.com/redis/go-redis/v9"
)

type Session interface {
	Set(ctx context.Context, key string, val string, duration time.Duration) error
	Find(ctx context.Context, key string) (string, error)
	RenewAndDelete(ctx context.Context, oldKey, newKey, newVal string, duration time.Duration) error
	Delete(ctx context.Context, key string) error
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

func (s *sessionImpl) Find(ctx context.Context, key string) (string, error) {
	val, err := s.redis.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", types.ErrNoData
	} else if err != nil {
		return "", errors.New(err)
	}

	return val, nil
}

func (s *sessionImpl) RenewAndDelete(ctx context.Context, oldKey, newKey, newVal string, duration time.Duration) error {
	pipe := s.redis.Pipeline()

	pipe.Del(ctx, oldKey)
	pipe.Set(ctx, newKey, newVal, duration)

	if _, err := pipe.Exec(ctx); err != nil {
		return errors.New(err)
	}

	return nil
}

func (s *sessionImpl) Delete(ctx context.Context, key string) error {
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return errors.New(err)
	}

	return nil
}
