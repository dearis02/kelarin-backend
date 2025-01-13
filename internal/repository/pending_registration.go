package repository

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type PendingRegistration interface {
	Set(ctx context.Context, key string, userID uuid.UUID) error
	IsExists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, key string) error
}

type pendingRegistrationImpl struct {
	redis *redis.Client
}

func NewPendingRegistration(redis *redis.Client) PendingRegistration {
	return &pendingRegistrationImpl{redis: redis}
}

func (r *pendingRegistrationImpl) Set(ctx context.Context, key string, userID uuid.UUID) error {
	if err := r.redis.Set(ctx, key, userID.String(), 0).Err(); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *pendingRegistrationImpl) IsExists(ctx context.Context, key string) (bool, error) {
	_, err := r.redis.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, errors.New(err)
	}

	return true, nil
}

func (r *pendingRegistrationImpl) Delete(ctx context.Context, key string) error {
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return errors.New(err)
	}

	return nil
}
