//go:build wireinject
// +build wireinject

package main

import (
	"kelarin/internal/config"
	"kelarin/internal/middleware"
	"kelarin/internal/provider"
	"kelarin/internal/queue/task"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"firebase.google.com/go/messaging"
	"github.com/alexliesenfeld/opencage"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/redis/go-redis/v9"
)

func newServer(db *sqlx.DB, esDB *elasticsearch.TypedClient, config *config.Config, redis *redis.Client, s3UploadManager *manager.Uploader, queueClient *asynq.Client, s3Client *s3.Client, s3PresignClient *s3.PresignClient, opencageClient *opencage.Client, firebaseMessagingClient *messaging.Client, midtransSnapClient *snap.Client, wsUpgrader *websocket.Upgrader, wsHub *types.WsHub, mainDBTx dbUtil.SqlxTx) (*provider.Server, error) {
	wire.Build(
		middleware.NewAuth,
		task.NewTempFile,
		provider.RepositorySet,
		provider.ServiceSet,
		provider.HandlerSet,
		provider.NewServer,
	)

	return &provider.Server{}, nil
}
