package main

import (
	"context"
	"log"
	"net/http"
	"os"

	apphttp "idempotent-payment/internal/http"
	"idempotent-payment/internal/payment"
	"idempotent-payment/internal/storage/postgres"
)

func main() {
	ctx := context.Background()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL is not set")
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

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
