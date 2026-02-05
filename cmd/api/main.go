package main

import (
	"context"
	"net/http"
	"os"

	"initial-airport-management-system/internal/airportops"
	"initial-airport-management-system/internal/booking"
	"initial-airport-management-system/internal/flight"
	"initial-airport-management-system/internal/passenger"
	"initial-airport-management-system/internal/platform/database"
	"initial-airport-management-system/internal/platform/logger"
)

func main() {
	log := logger.New("dev", "DEBUG")
	ctx := context.Background()

	// Database
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

	// Initialize Repositories
	flightRepo := flight.NewRepository(pool)
	bookingRepo := booking.NewRepository(pool)
	passengerRepo := passenger.NewRepository(pool)
	opsRepo := airportops.NewRepository(pool)

	// Initialize Services
	flightService := flight.NewService(flightRepo, log)
	bookingService := booking.NewService(bookingRepo, flightRepo, txManager, log)
	passengerService := passenger.NewService(passengerRepo, log)
	opsService := airportops.NewService(opsRepo, log)

	// Initialize Handlers
	flightHandler := flight.NewHandler(flightService)
	bookingHandler := booking.NewHandler(bookingService)
	passengerHandler := passenger.NewHandler(passengerService)
	opsHandler := airportops.NewHandler(opsService)

	// Router (Standard Mux)
	router := http.NewServeMux()

	// Flight Routes
	router.HandleFunc("POST /flights", flightHandler.Create)
	router.HandleFunc("GET /flights", flightHandler.Search)
	router.HandleFunc("POST /flights/status", flightHandler.UpdateStatus)

	// Booking Routes
	router.HandleFunc("POST /bookings", bookingHandler.Create)
	router.HandleFunc("GET /bookings/details", bookingHandler.GetDetails)
	router.HandleFunc("POST /bookings/cancel", bookingHandler.Cancel)

	// Passenger Routes
	router.HandleFunc("POST /passengers", passengerHandler.Create)
	router.HandleFunc("GET /passengers", passengerHandler.Get)

	// Ops Routes
	router.HandleFunc("POST /gates", opsHandler.CreateGate)
	router.HandleFunc("POST /gates/assign", opsHandler.AssignGate)

	// Start Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Info("Server starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Error("Server failed", "error", err)
	}
}
