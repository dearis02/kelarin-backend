package main

import (
	"context"
	"kelarin/internal/config"
	"kelarin/internal/queue"
	"kelarin/internal/types"
	awsUtil "kelarin/internal/utils/aws"
	dbUtil "kelarin/internal/utils/dbutil"
	firebaseUtil "kelarin/internal/utils/firebase_util"
	ws "kelarin/internal/utils/websocket"
	"kelarin/pkg/cron"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.NewApp("config/config.yaml")
	config.NewLogger(cfg)

	cron := cron.New()

	db, err := dbUtil.NewPostgres(&cfg.DataBase)
	if err != nil {
		log.Fatal().Caller().Err(err).Send()
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

	cronApp := newCronjob(db, mainDBTx, es, cfg, redis, queueClient, s3Client, s3Uploader, s3PresignClient, firebaseMessagingClient, wsUpgrader, wsHub)

	ctx := context.Background()

	for _, job := range cfg.Jobs {
		switch job.Name {
		case types.CronjobMarkOfferAsExpired:
			err = cron.RegisterJob(ctx, job, cronApp.OfferService.TaskMarkAsExpired)
			if err != nil {
				log.Fatal().Err(err).Send()
			}
		case types.CronjobUpdateOrderStatus:
			err = cron.RegisterJob(ctx, job, cronApp.OrderService.TaskUpdateOrderStatus)
			if err != nil {
				log.Fatal().Err(err).Send()
			}
		default:
			log.Fatal().Msgf("Unknown job name: %s", job.Name)
		}
	}

	go func() {
		log.Info().Msg("cronjob runner started")
		cron.Start()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Info().Msg("Shutting down cronjob runner...")

	err = cronApp.Close()
	if err != nil {
		log.Error().Err(err).Msg("Failed to close cronjob application")
	}

	log.Info().Msg("Cronjob runner shuted down successfully")
}
