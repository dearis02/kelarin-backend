package main

import (
	"context"
	"kelarin/internal/config"
	"kelarin/internal/middleware"
	"kelarin/internal/routes"
	dbUtil "kelarin/internal/utils/dbutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.NewAppConfig()
	logger := config.NewLogger(cfg)

	gin.SetMode(cfg.Mode())
	g := gin.New()
	g.Use(gin.RecoveryWithWriter(&logger))
	g.Use(gin.LoggerWithWriter(&logger))
	g.Use(middleware.HttpErrorHandler)

	db, err := dbUtil.NewPostgres(&cfg.DataBase)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	redis, err := dbUtil.NewRedisClient(cfg)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	server, err := newServer(db, cfg, redis)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	// Init routes region

	authRoutes := routes.NewAuth(g, server.AuthHandler)
	userRoutes := routes.NewUser(g, server.UserHandler)

	// End init routes region

	// Register routes

	authRoutes.Register()
	userRoutes.Register()

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
