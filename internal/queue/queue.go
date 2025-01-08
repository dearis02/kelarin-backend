package queue

import (
	"kelarin/internal/config"
	"kelarin/internal/types"
	workerUtil "kelarin/internal/utils/worker_util"

	"github.com/go-errors/errors"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

func NewAsynq(cfg *config.RedisConfig) (*asynq.Client, error) {
	addr, err := asynq.ParseRedisURI(cfg.ConString)
	if err != nil {
		return nil, errors.New(err)
	}

	client := asynq.NewClient(addr)

	err = client.Ping()
	if err != nil {
		return nil, errors.New(err)
	}

	return client, nil
}

func NewServer(cfg *config.Config, logger *workerUtil.WorkerLogger) *asynq.Server {
	addr, err := asynq.ParseRedisURI(cfg.Redis.ConString)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	queuePriorityMap := types.GetQueuePriorityNameMap(cfg.Environment)

	server := asynq.NewServer(addr, asynq.Config{
		Concurrency:  2,
		Queues:       queuePriorityMap,
		Logger:       logger,
		ErrorHandler: workerUtil.NewErrorHandler(logger),
	})

	return server
}
