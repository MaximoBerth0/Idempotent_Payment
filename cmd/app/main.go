package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	apphttp "idempotent-payment/internal/http"
	"idempotent-payment/internal/idempotency"
	"idempotent-payment/internal/logger"
	"idempotent-payment/internal/payment"
	"idempotent-payment/internal/product"
	"idempotent-payment/internal/storage/postgres"
	"idempotent-payment/internal/telemetry"
)

func main() {
	ctx := context.Background()
	log := logger.New()

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

	// Telemetry — must be set up before anything else so the
	// global TracerProvider is ready when other packages call otel.Tracer()
	shutdownTracer, err := telemetry.Setup(ctx, "idempotent-payment")
	if err != nil {
		log.Error("failed to setup telemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracer(shutdownCtx); err != nil {
			log.Error("failed to shutdown tracer", "error", err)
		}
	}()

	// Database
	pool, err := postgres.NewPool(ctx, connString)
	if err != nil {
		log.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	log.Info("database connected successfully")

	// Repository
	paymentRepo := postgres.NewPaymentRepository(pool, log, pool)
	productRepo := postgres.NewProductRepository(pool, log)
	idempotencyRepo := postgres.NewIdempotencyRepository(pool, log)

	// Service
	paymentService := payment.NewService(paymentRepo, log, productRepo)
	productService := product.NewService(productRepo, log)
	idempotencyService := idempotency.NewService(idempotencyRepo, log, func(ctx context.Context, fn func(context.Context) error) error {
		return postgres.WithTransaction(ctx, pool, fn)
	})

	// Handler
	paymentHandler := apphttp.NewPaymentHandler(paymentService, idempotencyService, log)
	productHandler := apphttp.NewProductHandler(productService, log)

	// Router
	router := apphttp.NewRouter(paymentHandler, productHandler)

	// Wrap the router with OTel middleware, this automatically creates a
	// root span for every incoming request and propagates trace context
	handler := otelhttp.NewHandler(router, "idempotent-payment",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)

	log.Info("server starting", "port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}
