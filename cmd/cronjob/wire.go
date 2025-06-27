//go:build wireinject
// +build wireinject

package main

import (
	"kelarin/internal/config"
	"kelarin/internal/provider"
	"kelarin/internal/queue/task"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"firebase.google.com/go/messaging"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func newCronjob(
	db *sqlx.DB,
	mainDBTx dbUtil.SqlxTx,
	esDB *elasticsearch.TypedClient,
	config *config.Config,
	redis *redis.Client,
	queueClient *asynq.Client,
	s3Client *s3.Client,
	s3UploadManager *manager.Uploader,
	s3PresignClient *s3.PresignClient,
	firebaseMessagingClient *messaging.Client,
	wsUpgrader *websocket.Upgrader,
	wsHub *types.WsHub,
) *provider.Cronjob {
	wire.Build(
		task.NewTempFile,
		provider.TaskRepositorySet,
		provider.TaskServiceSet,
		provider.NewCronjob,
	)

	return &provider.Cronjob{}
}
