package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	apphttp "idempotent-payment/internal/http"
	"idempotent-payment/internal/payment"
	"idempotent-payment/internal/storage/postgres"
)

func main() {
	ctx := context.Background()

	// Load .env file (only for local development)
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system environment variables")
	}

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Database
	pool, err := postgres.NewPool(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	log.Println("database connected successfully")

	// Repository
	paymentRepo := postgres.NewPaymentRepository(pool)

	// Service
	paymentService := payment.NewService(paymentRepo)

	// Handler
	paymentHandler := payment.NewHandler(paymentService)

	// Router
	router := apphttp.NewRouter(apphttp.Handlers{
		Health:        paymentHandler.Health,
		CreatePayment: paymentHandler.Create,
		GetPayment:    paymentHandler.GetByID,
	})

	log.Printf("Server running on :%s", port)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
