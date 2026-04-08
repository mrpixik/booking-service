package server

import (
	"net/http"

	_ "github.com/avito-internships/test-backend-1-mrpixik/docs"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters/logger"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/auth"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/middleware"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/middleware/rate_limiter"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type rateLimiter interface {
	Allow() bool
}

type jwtManager interface {
	GenerateToken(userID, role string) (string, error)
	ParseToken(tokenStr string) (*auth.Claims, error)
}

type authHandler interface {
	DummyLogin(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

type roomHandler interface {
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
}

type scheduleHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
}

type slotHandler interface {
	List(w http.ResponseWriter, r *http.Request)
}

type bookingHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	ListMy(w http.ResponseWriter, r *http.Request)
	Cancel(w http.ResponseWriter, r *http.Request)
}

type RouterConfig struct {
	Log      logger.LoggerAdapter
	Limiter  rateLimiter
	JWT      jwtManager
	Auth     authHandler
	Room     roomHandler
	Schedule scheduleHandler
	Slot     slotHandler
	Booking  bookingHandler
}

func InitRouter(cfg RouterConfig) chi.Router {

	router := chi.NewRouter()

	router.Use(
		middleware.WithLogging(cfg.Log),
		rate_limiter.WithRateLimiter(cfg.Limiter, cfg.Log),
	)

	router.Post("/dummyLogin", cfg.Auth.DummyLogin)
	router.Post("/register", cfg.Auth.Register)
	router.Post("/login", cfg.Auth.Login)
	// @Summary Health check
	// @Description Проверка доступности сервиса
	// @Tags system
	// @Success 200
	// @Router /_info [get]
	router.Get("/_info", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware(cfg.JWT))

		// admin + user
		r.Get("/rooms/list", cfg.Room.List)
		r.Get("/rooms/{roomId}/slots/list", cfg.Slot.List)

		// admin
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Post("/rooms/create", cfg.Room.Create)
			r.Post("/rooms/{roomId}/schedule/create", cfg.Schedule.Create)
			r.Get("/bookings/list", cfg.Booking.List)
		})

		// user
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireUser)
			r.Post("/bookings/create", cfg.Booking.Create)
			r.Get("/bookings/my", cfg.Booking.ListMy)
			r.Post("/bookings/{bookingId}/cancel", cfg.Booking.Cancel)
		})

	})

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	return router
}
