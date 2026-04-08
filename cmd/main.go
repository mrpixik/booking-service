// @title Booking Service API
// @version 1.0
// @description API для бронирования переговорных комнат

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters/logger"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters/logger/sl"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/auth"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/client/conference"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/config"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/middleware/rate_limiter"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/server"
	auth2 "github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/server/handlers/auth"
	booking3 "github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/server/handlers/booking"
	room3 "github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/server/handlers/room"
	schedule3 "github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/server/handlers/schedule"
	slot3 "github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/server/handlers/slot"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/booking"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/postgres"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/room"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/schedule"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/slot"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/user"
	booking2 "github.com/avito-internships/test-backend-1-mrpixik/internal/service/booking"
	room2 "github.com/avito-internships/test-backend-1-mrpixik/internal/service/room"
	schedule2 "github.com/avito-internships/test-backend-1-mrpixik/internal/service/schedule"
	slot2 "github.com/avito-internships/test-backend-1-mrpixik/internal/service/slot"
	user2 "github.com/avito-internships/test-backend-1-mrpixik/internal/service/user"
)

func main() {

	cfg := config.MustLoad()

	fmt.Printf("%+v\n", cfg)

	// LOGGER
	log := sl.NewSlogLogger(cfg.Env)
	log.Info("logger initialized")

	// GS context
	ctxApp, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// context для бд, отменяется после завершения сервера
	ctxDB, cancelDB := context.WithCancel(context.Background())
	defer cancelDB()

	// DB connection
	pool, err := postgres.ConnectPostgres(ctxDB, cfg.Postgres, cfg.Env)
	if err != nil {
		log.Error("connect database " + err.Error())
		os.Exit(1)
	}
	log.Info("connection with database initialized")
	defer func() {
		pool.Close()
		log.Info("connection with database closed")
	}()

	// Transaction Manager
	txManager := postgres.NewTransactionManager(pool)

	//JWT
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiration)

	// Fake conference client
	confClient, confCleanup, err := conference.NewBufconnClient()
	if err != nil {
		log.Error("create conference client " + err.Error())
		os.Exit(1)
	}
	log.Info("conference client initialized")
	// Repository
	userRepo := user.NewUserRepository(pool)
	roomRepo := room.NewRoomRepository(pool)
	scheduleRepo := schedule.NewScheduleRepository(pool)
	slotRepo := slot.NewSlotRepository(pool)
	bookRepo := booking.NewBookingRepository(pool)

	//Service
	authService := user2.NewAuthService(userRepo, jwtManager)
	roomService := room2.NewRoomService(roomRepo)
	scheduleService := schedule2.NewScheduleService(scheduleRepo, roomRepo, txManager)
	slotService := slot2.NewSlotService(slotRepo, scheduleRepo, roomRepo, txManager)
	bookService := booking2.NewBookingService(bookRepo, slotRepo, txManager, confClient)

	// Handler
	authHandler := auth2.NewAuthHandler(jwtManager, authService)
	roomHandler := room3.NewRoomHandler(roomService)
	scheduleHandler := schedule3.NewScheduleHandler(scheduleService)
	slotHandler := slot3.NewSlotHandler(slotService)
	bookHandler := booking3.NewBookingHandler(bookService)

	// Rate limiter
	tokenBucketLimiter := rate_limiter.NewTokenBucket(cfg.HTTP.RateLimiter.MaxRPC, cfg.HTTP.RateLimiter.RPCRefill)

	// ROUTER & SERVER
	r := server.InitRouter(server.RouterConfig{
		Log:      log,
		Limiter:  tokenBucketLimiter,
		JWT:      jwtManager,
		Auth:     authHandler,
		Room:     roomHandler,
		Schedule: scheduleHandler,
		Slot:     slotHandler,
		Booking:  bookHandler,
	})

	srv := &http.Server{
		//Addr:    ":" + cfg.HTTP.Port,
		Addr:    ":" + cfg.HTTP.Port,
		Handler: r,
		BaseContext: func(net.Listener) context.Context {
			return ctxApp
		},
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.ShutdownTimeout,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server start up: %s", err.Error())
		}
	}()
	log.Info("listening on: " + cfg.HTTP.Port)

	gracefulShutdown(ctxApp, cfg.HTTP, log, srv, cancelDB, confCleanup)
}

func gracefulShutdown(ctxApp context.Context, cfg config.HTTPServer, log logger.LoggerAdapter, srv *http.Server, cancelDB context.CancelFunc, confCleanup func()) {
	<-ctxApp.Done()
	log.Info("shutdown signal received. starting graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Отключение клиента
	confCleanup()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown: %s", err.Error())
	} else {
		log.Info("server gracefully stopped")
	}

	cancelDB()
}
