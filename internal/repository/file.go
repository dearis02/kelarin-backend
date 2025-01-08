package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/redis/go-redis/v9"
)

type File interface {
	SetTemp(ctx context.Context, req []types.FileTemp) error
	GetTemp(ctx context.Context, fileName string) (types.FileGetTempRes, error)
}

type fileUploadImpl struct {
	redis *redis.Client
}

func NewFile(redis *redis.Client) File {
	return &fileUploadImpl{redis}
}

func (r *fileUploadImpl) SetTemp(ctx context.Context, req []types.FileTemp) error {
	pipeline := r.redis.Pipeline()

	for _, file := range req {
		pipeline.SetEx(ctx, types.GetUploadedFileKey(file.Name), file.Name, file.Expiration)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *fileUploadImpl) GetTemp(ctx context.Context, fileName string) (types.FileGetTempRes, error) {
	res := types.FileGetTempRes{}
	key := types.GetUploadedFileKey(fileName)

	err := r.redis.HGetAll(ctx, key).Scan(&res)
	if errors.Is(err, redis.Nil) {
		return res, types.ErrNoData
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
