package main

import (
	"context"
	"net/http"
	"os"

	"initial-airport-management-system/internal/booking"
	"initial-airport-management-system/internal/flight"
	"initial-airport-management-system/internal/platform/database"
	"initial-airport-management-system/internal/platform/logger"
)

func main() {
	// 1. Infrastructure
	log := logger.New("dev", "DEBUG")
	ctx := context.Background()

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

	// Transaction Manager
	txManager := database.NewTxManager(pool)

	// 2. Wiring Layers (Repo -> Service -> Handler)

	// Flight Module
	flightRepo := flight.NewRepository(pool)
	flightService := flight.NewService(flightRepo, log)
	flightHandler := flight.NewHandler(flightService)

	// Booking Module (Needs FlightRepo + TxManager)
	bookingRepo := booking.NewRepository(pool)
	bookingService := booking.NewService(bookingRepo, flightRepo, txManager, log)
	bookingHandler := booking.NewHandler(bookingService)

	// 3. Routing (Go 1.22 Standard Library)
	router := http.NewServeMux()

	// Routes
	router.HandleFunc("POST /flights", flightHandler.Create)
	router.HandleFunc("GET /flights", flightHandler.Search)
	router.HandleFunc("POST /bookings", bookingHandler.Create)

	// 4. Server Start
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Info("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Error("Server failed", "error", err)
	}
}
