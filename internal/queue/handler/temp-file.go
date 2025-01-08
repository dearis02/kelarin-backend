package taskHandler

import (
	"context"
	"encoding/json"
	"kelarin/internal/types"
	pkg "kelarin/pkg/utils"
	"os"

	"github.com/go-errors/errors"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type QueueTempFile interface {
	DeleteTempFile(ctx context.Context, t *asynq.Task) error
}

type tempFileImpl struct{}

func NewQueueTempFile() QueueTempFile {
	return &tempFileImpl{}
}

func (tempFileImpl) DeleteTempFile(ctx context.Context, t *asynq.Task) error {
	payload := types.QueueDeleteTempFilePayload{}

	err := json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		return errors.New(err)
	}

	if pkg.IsFileExist(payload.FilePath) {
		log.Info().Msgf("Deleting temp file: %s", payload.FilePath)
		err := os.Remove(payload.FilePath)
		if err != nil {
			return errors.New(err)
		}

		log.Info().Msgf("Temp file deleted: %s", payload.FilePath)
	}

	return nil
}
