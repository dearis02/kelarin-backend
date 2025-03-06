package main

import (
	"context"
	"encoding/json"
	"fmt"
	"kelarin/internal/config"
	"kelarin/internal/middleware"
	"kelarin/internal/queue"
	"kelarin/internal/routes"
	"kelarin/internal/types"
	"kelarin/internal/utils"
	awsUtil "kelarin/internal/utils/aws"
	dbUtil "kelarin/internal/utils/dbutil"
	fileSystemUtil "kelarin/internal/utils/file_system"
	firebaseUtil "kelarin/internal/utils/firebase_util"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexliesenfeld/opencage"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// type province map[string]string

func main() {
	cfg := config.NewApp()
	logger := config.NewLogger(cfg)
	if err := fileSystemUtil.InitTempDir(); err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	gin.SetMode(cfg.Mode())
	g := gin.New()
	g.Use(gin.RecoveryWithWriter(&logger))
	g.Use(gin.LoggerWithWriter(&logger))

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.Server.CORS.AllowedOrigins
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Time-Zone"}
	g.Use(cors.New(corsConfig))

	db, err := dbUtil.NewPostgres(&cfg.DataBase)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	redis, err := dbUtil.NewRedisClient(cfg)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	es, err := dbUtil.NewElasticsearchClient(cfg)
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

	openCageClient := opencage.New(cfg.OpenCageApiKey)

	authMiddleware := middleware.NewAuth(cfg)

	firebaseApp := firebaseUtil.NewApp(cfg)
	firebaseMessagingClient := firebaseUtil.NewMessagingClient(firebaseApp)

	midtransSnapClient := utils.NewMidtransSnapClient(cfg.Midtrans.ServerKey, cfg.Midtrans.Env(), cfg.Midtrans.NotificationURL)

	server, err := newServer(db, es, cfg, redis, s3Uploader, queueClient, s3Client, s3PresignClient, openCageClient, authMiddleware, firebaseMessagingClient, midtransSnapClient)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

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

	// End init routes region

	// Register routes

	authRoutes.Register()
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

	// register websocket
	wsClient := &WSClient{
		Conns: make(map[string]*websocket.Conn),
	}
	g.GET("/ws", func(c *gin.Context) {
		fmt.Println(c.GetQuery("token"))
		token, exs := c.GetQuery("token")
		if !exs {
			log.Error().Msg("token not found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("web_socket_auth", token)

	}, webSocketHandler(wsClient))

	// End routes registration

	// CODE: insert provinces data

	// provinceRepo := repository.NewProvince(db)

	// provinceData, err := os.Open("provinsi.json")
	// if err != nil {
	// 	log.Fatal().Err(err).Send()
	// }

	// decoder := json.NewDecoder(provinceData)

	// var provinces province
	// err = decoder.Decode(&provinces)
	// if err != nil {
	// 	log.Fatal().Err(err).Send()
	// }

	// provincesData := []types.Province{}
	// for key, val := range provinces {
	// 	id, err := strconv.ParseInt(key, 10, 64)
	// 	if err != nil {
	// 		log.Fatal().Err(err).Send()
	// 	}

	// 	provincesData = append(provincesData, types.Province{
	// 		ID:   id,
	// 		Name: val,
	// 	})
	// }

	// err = provinceRepo.Create(context.Background(), provincesData)
	// if err != nil {
	// 	log.Fatal().Err(err).Send()
	// }

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

func webSocketHandler(client *WSClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsUp := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// allow all origins for development purposes
				return true
			},
			HandshakeTimeout: 5 * time.Second,
			ReadBufferSize:   1024,
			WriteBufferSize:  1024,
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
				log.Error().Err(reason).Msg("websocket error")
			},
		}

		_userID, exists := c.Get("web_socket_auth")
		if !exists {
			c.Error(errors.New(types.AppErr{Code: http.StatusUnauthorized, Err: errors.New("web socket middleware not used")}))
		}

		userID := _userID.(string)

		con, err := wsUp.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Fatal().Err(err).Msg("websocket error")
		}

		defer con.Close()

		// store connection
		client.Lock()
		client.Conns[userID] = con
		client.Unlock()
		fmt.Println("user id", userID)

		for {
			// msgType, msg, err
			t, msg, err := con.ReadMessage()
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Info().Msg("client closed connection")
				continue
			} else if err != nil {
				log.Error().Err(err).Msg("failed to read message")
				continue
			}

			fmt.Println("message type", t)

			data := Data{}
			err = json.Unmarshal(msg, &data)
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}

			fmt.Println(data)

			// find user on based on room id except the sender

			// send message to target
			client.Lock()
			targetCon, exs := client.Conns[data.TargetID]
			if !exs {
				continue
			}
			client.Unlock()

			err = targetCon.WriteMessage(websocket.TextMessage, []byte(data.Message))
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}
		}
	}
}

type Data struct {
	TargetID string `json:"target_id"`
	Message  string `json:"message"`
}

type WSClient struct {
	sync.RWMutex
	Conns map[string]*websocket.Conn
}
