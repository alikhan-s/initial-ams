package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"initial-airport-management-system/internal/airportops"
	"initial-airport-management-system/internal/booking"
	"initial-airport-management-system/internal/flight"
	"initial-airport-management-system/internal/passenger"
	"initial-airport-management-system/internal/platform/database"
	"initial-airport-management-system/internal/platform/logger"
)

func main() {
	// Initializing Logger
	log := logger.New("dev", "DEBUG")
	slog.SetDefault(log)

	ctx := context.Background()

	// Database Connection
	dbConfig := database.Config{
		Host: "localhost", Port: "5432", User: "postgres",
		Password: "123456", DBName: "airport_db", SSLMode: "disable",
	}
	pool, err := database.New(ctx, dbConfig)
	if err != nil {
		log.Error("cannot connect to db", "err", err)
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing database connection pool...")
		pool.Close()
		log.Info("Database connection pool closed")
	}()

	txManager := database.NewTxManager(pool)

	// Initializing Repositories
	flightRepo := flight.NewRepository(pool)
	bookingRepo := booking.NewRepository(pool)
	passengerRepo := passenger.NewRepository(pool)
	opsRepo := airportops.NewRepository(pool)

	// Initializing Services
	flightService := flight.NewService(flightRepo, log)
	bookingService := booking.NewService(bookingRepo, flightRepo, txManager, log)
	passengerService := passenger.NewService(passengerRepo, log)
	opsService := airportops.NewService(opsRepo, log)

	// Initializing Handlers
	flightHandler := flight.NewHandler(flightService)
	bookingHandler := booking.NewHandler(bookingService)
	passengerHandler := passenger.NewHandler(passengerService)
	opsHandler := airportops.NewHandler(opsService)

	// Setup Router
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	v1 := router.Group("/api/v1")
	{
		// FLIGHTS
		flights := v1.Group("/flights")
		{
			flights.POST("", flightHandler.Create)
			flights.GET("", flightHandler.Search)
			flights.PATCH("/:id/status", flightHandler.UpdateStatus)
		}

		// BOOKINGS
		bookings := v1.Group("/bookings")
		{
			bookings.POST("", bookingHandler.Create)
			bookings.GET("/:id", bookingHandler.GetDetails)
			bookings.POST("/:id/cancel", bookingHandler.Cancel)
		}

		// PASSENGERS
		passengers := v1.Group("/passengers")
		{
			passengers.POST("", passengerHandler.Create)
			passengers.GET("/:id", passengerHandler.Get)
		}

		// GATES (OPS)
		gates := v1.Group("/gates")
		{
			gates.POST("", opsHandler.CreateGate)
			gates.POST("/assign", opsHandler.AssignGate)
		}
	}

	// Server Configuration
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Info("Server starting on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server exiting")
}
