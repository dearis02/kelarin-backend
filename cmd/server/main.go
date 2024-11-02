package main

import (
	"context"
	"kelarin/internal/config"
	"kelarin/internal/handler"
	"kelarin/internal/middleware"
	"kelarin/internal/routes"
	"kelarin/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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

	// Repository region
	// End repository region

	// Service region
	userService := service.NewUser()
	// End service region

	// Handler region
	userHandler := handler.NewUser(userService)
	// End handler region

	// Routes region
	userRoutes := routes.NewUser(g, userHandler)
	// End routes region

	// Register routes
	userRoutes.Register()

	startServer(g, cfg)
}

func startServer(g *gin.Engine, cfg *config.Config) {
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

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server shuted down gracefully")
}
