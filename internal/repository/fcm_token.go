package repository

import (
	"context"
	"kelarin/internal/types"
	"time"

	"github.com/go-errors/errors"
	"github.com/redis/go-redis/v9"
)

type FCMToken interface {
	Find(ctx context.Context, key string) (string, error)
	Save(ctx context.Context, key, token string, expiration time.Duration) error
}

type fcmTokenImpl struct {
	redisDB *redis.Client
}

func NewFCMToken(redisDB *redis.Client) FCMToken {
	return &fcmTokenImpl{
		redisDB: redisDB,
	}
}

func (r *fcmTokenImpl) Find(ctx context.Context, key string) (string, error) {
	res := ""

	err := r.redisDB.Get(ctx, key).Scan(&res)
	if errors.Is(err, redis.Nil) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *fcmTokenImpl) Save(ctx context.Context, key, token string, expiration time.Duration) error {
	if err := r.redisDB.Set(ctx, key, token, expiration).Err(); err != nil {
		return errors.New(err)
	}

	return nil
}
