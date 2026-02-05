package main

import (
	"context"
	"log/slog"
	"os"

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
	defer pool.Close()
	txManager := database.NewTxManager(pool)

	// Initializing Repositories
	flightRepo := flight.NewRepository(pool)
	bookingRepo := booking.NewRepository(pool)
	passengerRepo := passenger.NewRepository(pool)
	opsRepo := airportops.NewRepository(pool)

	// Initializing Services
	flightService := flight.NewService(flightRepo, log)
	// Booking Service требует flightRepo для проверки рейсов и txManager для транзакций
	bookingService := booking.NewService(bookingRepo, flightRepo, txManager, log)
	passengerService := passenger.NewService(passengerRepo, log)
	opsService := airportops.NewService(opsRepo, log)

	// Initializing Handlers
	flightHandler := flight.NewHandler(flightService)
	bookingHandler := booking.NewHandler(bookingService)
	passengerHandler := passenger.NewHandler(passengerService)
	opsHandler := airportops.NewHandler(opsService)

	router := gin.Default()

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

	log.Info("Server starting on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Error("Server failed", "error", err)
	}
}
