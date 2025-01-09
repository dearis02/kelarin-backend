package main

import (
	"context"
	"kelarin/internal/config"
	"kelarin/internal/middleware"
	"kelarin/internal/queue"
	"kelarin/internal/routes"
	awsUtil "kelarin/internal/utils/aws"
	dbUtil "kelarin/internal/utils/dbutil"
	fileSystemUtil "kelarin/internal/utils/file_system"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.NewAppConfig()
	logger := config.NewLogger(cfg)
	if err := fileSystemUtil.InitTempDir(); err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	gin.SetMode(cfg.Mode())
	g := gin.New()
	g.Use(gin.RecoveryWithWriter(&logger))
	g.Use(gin.LoggerWithWriter(&logger))
	g.Use(middleware.HttpErrorHandler)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.Server.CORS.AllowedOrigins
	g.Use(cors.New(corsConfig))

	db, err := dbUtil.NewPostgres(&cfg.DataBase)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	redis, err := dbUtil.NewRedisClient(cfg)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	queueClient, err := queue.NewAsynq(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to queue")
	}

	s3Config := awsUtil.NewConfig()
	s3Client := awsUtil.NewS3ClientFromConfig(s3Config)
	s3Uploader := manager.NewUploader(s3Client)
	s3PresignClient := awsUtil.NewS3PresignClient(s3Client)

	server, err := newServer(db, cfg, redis, s3Uploader, queueClient, s3Client, s3PresignClient)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	authMiddleware := middleware.NewAuth(cfg)

	// Init routes region

	authRoutes := routes.NewAuth(g, server.AuthHandler)
	userRoutes := routes.NewUser(g, server.UserHandler)
	fileRoutes := routes.NewFile(g, server.FileHandler)

	// End init routes region

	// Register routes

	authRoutes.Register()
	userRoutes.Register()
	fileRoutes.Register(authMiddleware)

	// End routes registration

	startServer(g, db, cfg)
}

func startServer(g *gin.Engine, db *sqlx.DB, cfg *config.Config) {
	srv := &http.Server{
		Addr:    cfg.Address(),
		Handler: g,
	}

	go func() {
		log.Info().Msg("server started on " + cfg.Address())

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info().Msg("Closing database connection...")
	err := dbUtil.ClosePostgresConnection(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to close database connection")
	}
	log.Info().Msg("Database connection closed")

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server shuted down gracefully")
}
