package main

import (
	"context"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	apphttp "idempotent-payment/internal/http"
	"idempotent-payment/internal/logger"
	"idempotent-payment/internal/payment"
	"idempotent-payment/internal/storage/postgres"
)

func main() {
	ctx := context.Background()

	log := logger.New()

	// Load .env file (only for local development)
	if err := godotenv.Load(); err != nil {
		log.Warn(".env file not found, using system environment variables")
	}

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Error("DATABASE_URL is not set")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Database
	pool, err := postgres.NewPool(ctx, connString)
	if err != nil {
		log.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	log.Info("database connected successfully")

	// Repository
	paymentRepo := postgres.NewPaymentRepository(pool)

	// Service
	paymentService := payment.NewService(paymentRepo, log)

	// Handler
	paymentHandler := payment.NewHandler(paymentService, log)

	// Router
	router := apphttp.NewRouter(apphttp.Handlers{
		Health:        paymentHandler.Health,
		CreatePayment: paymentHandler.Create,
		GetPayment:    paymentHandler.GetByID,
	})

	log.Info("server starting", "port", port)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}
