package api

import (
	"evently/internal/api/handler"
	"evently/internal/api/middleware"
	"evently/internal/data"
	"evently/internal/service"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
)

func NewRouter(db *pgxpool.Pool, logger *slog.Logger, jwtSecret string) http.Handler {
	r := chi.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	})
	r.Use(corsMiddleware.Handler)

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	userRepo := &data.UserRepository{DB: db}
	eventRepo := &data.EventRepository{DB: db}
	dataBookingRepo := &data.BookingRepository{DB: db}
	bookingRepoWithTx := &service.BookingRepositoryWithTx{DB: db, BookingRepository: dataBookingRepo}

	authService := service.NewAuthService(userRepo, jwtSecret)
	bookingService := service.NewBookingService(bookingRepoWithTx, logger)

	authHandler := handler.NewAuthHandler(authService, logger)
	eventHandler := handler.NewEventHandler(eventRepo)
	bookingHandler := handler.NewBookingHandler(bookingService, logger) // Changed this line

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	r.Route("/events", func(r chi.Router) {
		r.Get("/", eventHandler.ListEvents)
		r.Get("/{id}", eventHandler.GetEvent)
		r.With(middleware.JWTAuth(jwtSecret)).Post("/{id}/book", bookingHandler.CreateBooking)
	})

	// Inside the NewRouter function, change the /bookings route
	r.Route("/bookings", func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtSecret))
		r.Get("/", bookingHandler.GetUserBookings)
		r.Post("/{id}/cancel", bookingHandler.CancelBooking) // Change this line
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtSecret))
		r.Use(middleware.AdminOnly)
		r.Post("/events", eventHandler.CreateEvent)
	})

	return r
}
