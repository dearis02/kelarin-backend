package main

import (
	"context"
	"kelarin/internal/config"
	"kelarin/internal/middleware"
	"kelarin/internal/queue"
	"kelarin/internal/routes"
	"kelarin/internal/utils"
	awsUtil "kelarin/internal/utils/aws"
	dbUtil "kelarin/internal/utils/dbutil"
	fileSystemUtil "kelarin/internal/utils/file_system"
	firebaseUtil "kelarin/internal/utils/firebase_util"
	ws "kelarin/internal/utils/websocket"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexliesenfeld/opencage"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// type province map[string]string

func main() {
	cfg := config.NewApp("config/config.yaml")
	logger := config.NewLogger(cfg)
	if err := fileSystemUtil.InitTempDir(); err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	gin.SetMode(cfg.Mode())
	g := gin.New()
	g.Use(gin.RecoveryWithWriter(&logger))

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.Server.CORS.AllowedOrigins
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Time-Zone"}
	corsConfig.ExposeHeaders = []string{"Content-Disposition"}
	g.Use(cors.New(corsConfig))

	db, err := dbUtil.NewPostgres(&cfg.DataBase)
	if err != nil {
		log.Fatal().Stack().Err(errors.New(err)).Send()
	}

	redis, err := dbUtil.NewRedisClient(cfg)
	if err != nil {
		log.Fatal().Stack().Err(errors.New(err)).Send()
	}

	es, err := dbUtil.NewElasticsearchClient(cfg.Elasticsearch)
	if err != nil {
		log.Fatal().Stack().Caller().Err(errors.New(err)).Send()
	}

	esPing, err := es.Ping().Do(context.Background())
	if err != nil {
		log.Fatal().Stack().Err(errors.New(err)).Msg("failed to ping elasticsearch")
	} else if !esPing {
		log.Fatal().Stack().Msg("elasticsearch is not available")
	}

	queueClient, err := queue.NewAsynq(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(errors.New(err)).Msg("Failed to connect to queue")
	}

	awsCfg := awsUtil.NewConfig()
	s3Client := awsUtil.NewS3ClientFromConfig(awsCfg)
	s3Uploader := manager.NewUploader(s3Client)
	s3PresignClient := awsUtil.NewS3PresignClient(s3Client)

	openCageClient := opencage.New(cfg.OpenCageApiKey)

	firebaseApp := firebaseUtil.NewApp(cfg)
	firebaseMessagingClient := firebaseUtil.NewMessagingClient(firebaseApp)

	midtransSnapClient := utils.NewMidtransSnapClient(cfg.Midtrans.ServerKey, cfg.Midtrans.Env(), cfg.Midtrans.NotificationURL)

	wsUpgrader := ws.NewWsUpgrader(cfg)
	wsHub := ws.NewWsHub()

	mainDBTx := dbUtil.NewSqlxTx(db)
	server, err := newServer(db, es, cfg, redis, s3Uploader, queueClient, s3Client, s3PresignClient, openCageClient, firebaseMessagingClient, midtransSnapClient, wsUpgrader, wsHub, mainDBTx)
	if err != nil {
		log.Fatal().Err(errors.New(err)).Send()
	}

	authMiddleware := server.AuthMiddleware

	// prometheus

	prom := prometheus.NewRegistry()
	promHttpTrafficMiddleware := middleware.NewPrometheusHttpTraffic(prom)
	reqTotal := promHttpTrafficMiddleware.RequestTotalCollector()
	reqDuration := promHttpTrafficMiddleware.RequestDurationCollector()

	if err := promHttpTrafficMiddleware.Register([]prometheus.Collector{reqTotal, reqDuration}); err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	// prometheus middleware must be registered first to get the response status code that set on http error handler middleware
	g.Use(middleware.RequestDuration(reqTotal, reqDuration))
	g.Use(middleware.HttpErrorHandler)
	g.GET("/metrics", gin.WrapH(promhttp.HandlerFor(prom, promhttp.HandlerOpts{})))

	// end prometheus

	// Init routes region

	authRoutes := routes.NewAuth(g, server.AuthHandler)
	userRoutes := routes.NewUser(g, server.UserHandler)
	fileRoutes := routes.NewFile(g, server.FileHandler)
	serviceProviderRoutes := routes.NewServiceProvider(g, server.ServiceProviderHandler)
	serviceRoutes := routes.NewService(g, server.ServiceHandler)
	provinceRoutes := routes.NewProvince(g, server.ProvinceHandler)
	cityRoutes := routes.NewCity(g, server.CityHandler)
	serviceCategoryRoutes := routes.NewServiceCategory(g, server.ServiceCategoryHandler)
	userAddressRoutes := routes.NewUserAddress(g, server.UserAddressHandler)
	offerRoutes := routes.NewOffer(g, server.OfferHandler)
	offerNegotiationRoutes := routes.NewOfferNegotiation(g, server.OfferNegotiationHandler)
	notificationRoutes := routes.NewNotification(g, server.NotificationHandler)
	paymentRoutes := routes.NewPayment(g, server.PaymentHandler)
	orderRoutes := routes.NewOrder(g, server.OrderHandler)
	paymentMethodRoutes := routes.NewPaymentMethod(g, server.PaymentMethodHandler)
	reportRoutes := routes.NewReport(g, server.ReportHandler)
	chatRoutes := routes.NewChat(g, server.ChatHandler)

	// End init routes region

	// Register routes

	authRoutes.Register(authMiddleware)
	userRoutes.Register()
	fileRoutes.Register(authMiddleware)
	serviceProviderRoutes.Register(authMiddleware)
	serviceRoutes.Register(authMiddleware)
	provinceRoutes.Register()
	cityRoutes.Register()
	serviceCategoryRoutes.Register(authMiddleware)
	userAddressRoutes.Register(authMiddleware)
	offerRoutes.Register(authMiddleware)
	offerNegotiationRoutes.Register(authMiddleware)
	notificationRoutes.Register(authMiddleware)
	paymentRoutes.Register(authMiddleware)
	orderRoutes.Register(authMiddleware)
	paymentMethodRoutes.Register()
	reportRoutes.Register(authMiddleware)
	chatRoutes.Register(authMiddleware)

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
		log.Error().Err(err).Msg("Failed to close database connection")
	}
	log.Info().Msg("Database connection closed")

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server shuted down gracefully")
}
