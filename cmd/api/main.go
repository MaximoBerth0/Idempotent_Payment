package main

import (
	"context"
	"log"
	"os"

	"idempotent-payment/internal/storage/postgres"
)

func main() {
	ctx := context.Background()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := postgres.NewPool(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	log.Println("Database connected successfully")
}
