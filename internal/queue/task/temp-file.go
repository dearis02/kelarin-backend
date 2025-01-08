package task

import (
	"context"
	"encoding/json"
	"kelarin/internal/types"
	"time"

	"github.com/go-errors/errors"
	"github.com/hibiken/asynq"
)

type TempFile interface {
	Delete(ctx context.Context, queueName string, req types.FileTemp, delay time.Duration) error
}

type fileUploadImpl struct {
	client *asynq.Client
}

func NewTempFile(client *asynq.Client) TempFile {
	return &fileUploadImpl{client}
}

func (r fileUploadImpl) Delete(ctx context.Context, queueName string, req types.FileTemp, delay time.Duration) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return errors.New(err)
	}

	task := asynq.NewTask(types.TaskDeleteTempFile, payload, asynq.Queue(queueName))

	_, err = r.client.Enqueue(task, asynq.ProcessIn(delay))
	if err != nil {
		return errors.New(err)
	}

	return nil
}
