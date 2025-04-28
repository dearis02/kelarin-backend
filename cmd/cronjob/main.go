package main

import (
	"context"
	"flag"
	"fmt"
	"kelarin/internal/config"
	"kelarin/internal/middleware"
	"kelarin/internal/queue"
	"kelarin/internal/types"
	awsUtil "kelarin/internal/utils/aws"
	dbUtil "kelarin/internal/utils/dbutil"
	firebaseUtil "kelarin/internal/utils/firebase_util"
	ws "kelarin/internal/utils/websocket"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var taskName string

func init() {
	flag.StringVar(&taskName, "task", "", "Run task")
	flag.Parse()
}

func main() {
	if taskName == "" {
		log.Fatal().Msg("task name required")
	}

	cfg := config.NewApp("config/config.yaml")
	logger := config.NewLogger(cfg)

	db, err := dbUtil.NewPostgres(&cfg.DataBase)
	if err != nil {
		logger.Fatal().Caller().Err(err).Send()
	}

	redis, err := dbUtil.NewRedisClient(cfg)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	es, err := dbUtil.NewElasticsearchClient(cfg.Elasticsearch)
	if err != nil {
		log.Fatal().Stack().Caller().Err(err).Send()
	}

	esPing, err := es.Ping().Do(context.Background())
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("failed to ping elasticsearch")
	} else if !esPing {
		log.Fatal().Stack().Msg("elasticsearch is not available")
	}

	queueClient, err := queue.NewAsynq(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to queue")
	}

	awsCfg := awsUtil.NewConfig()
	s3Client := awsUtil.NewS3ClientFromConfig(awsCfg)
	s3Uploader := manager.NewUploader(s3Client)
	s3PresignClient := awsUtil.NewS3PresignClient(s3Client)

	firebaseApp := firebaseUtil.NewApp(cfg)
	firebaseMessagingClient := firebaseUtil.NewMessagingClient(firebaseApp)

	wsUpgrader := ws.NewWsUpgrader(cfg)
	wsHub := ws.NewWsHub()

	mainDBTx := dbUtil.NewSqlxTx(db)
	authMiddleware := middleware.NewAuth(cfg)

	cronApp := newCronjob(db, mainDBTx, es, cfg, redis, queueClient, s3Client, s3Uploader, s3PresignClient, authMiddleware, firebaseMessagingClient, wsUpgrader, wsHub)

	close := func() {
		err := cronApp.Close()
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	}

	defer close()

	ctx := context.Background()

	now := time.Now()

	id := uuid.New()
	switch taskName {
	case types.CronjobMarkOfferAsExpired:
		log.Info().Str("id", id.String()).Str("task", taskName).Msgf("Task - %s started...", taskName)

		err = cronApp.OfferService.TaskMarkAsExpired(ctx)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}
	default:
		logger.Fatal().Msg("Unknown task")
	}

	elapsed := time.Since(now)

	log.Info().
		Str("id", id.String()).
		Str("task", taskName).
		Str("elapsed", fmt.Sprintf("%f seconds", elapsed.Seconds())).
		Msgf("Task - %s finished", taskName)
}
