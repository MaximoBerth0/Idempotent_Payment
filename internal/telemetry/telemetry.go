package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func Setup(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	// Merge your service attributes into the SDK default resource
	// (which already carries process, OS, runtime info, etc.)
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

/* db example
package postgres

import (
    "context"
    "fmt"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("idempotent-payment/postgres")

type PaymentRepository struct {
    pool *pgxpool.Pool
}

func (r *PaymentRepository) Save(ctx context.Context, p *payment.Payment) error {
    ctx, span := tracer.Start(ctx, "PaymentRepository.Save")
    defer span.End()

    span.SetAttributes(
        attribute.String("db.system", "postgresql"),
        attribute.String("db.operation", "INSERT"),
        attribute.String("db.table", "payments"),
        attribute.String("payment.id", p.ID.String()),
    )

    _, err := r.pool.Exec(ctx, `
        INSERT INTO payments (id, amount, currency, status)
        VALUES ($1, $2, $3, $4)
    `, p.ID, p.Amount, p.Currency, p.Status)

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return fmt.Errorf("save payment: %w", err)
    }

    return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
    ctx, span := tracer.Start(ctx, "PaymentRepository.GetByID")
    defer span.End()

    span.SetAttributes(
        attribute.String("db.system", "postgresql"),
        attribute.String("db.operation", "SELECT"),
        attribute.String("db.table", "payments"),
        attribute.String("payment.id", id.String()),
    )

    var p payment.Payment
    err := r.pool.QueryRow(ctx, `
        SELECT id, amount, currency, status FROM payments WHERE id = $1
    `, id).Scan(&p.ID, &p.Amount, &p.Currency, &p.Status)

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, fmt.Errorf("get payment by id: %w", err)
    }

    return &p, nil
}
*/

/* http example
package http

import (
    "encoding/json"
    "net/http"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("idempotent-payment/http")

type PaymentHandler struct {
    paymentService    payment.Service
    idempotencyService idempotency.Service
    log               *slog.Logger
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
    // otelhttp.NewHandler (in main.go) already created the root span —
    // here you add a child span with business-level detail
    ctx, span := tracer.Start(r.Context(), "PaymentHandler.CreatePayment")
    defer span.End()

    idempotencyKey := r.Header.Get("Idempotency-Key")
    span.SetAttributes(
        attribute.String("http.idempotency_key", idempotencyKey),
    )

    var req CreatePaymentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "invalid request body")
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    span.SetAttributes(
        attribute.String("payment.currency", req.Currency),
        attribute.Float64("payment.amount", req.Amount),
    )

    result, err := h.paymentService.ProcessPayment(ctx, req) // ctx carries the span
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        http.Error(w, "payment failed", http.StatusInternalServerError)
        return
    }

    span.SetStatus(codes.Ok, "")
    span.SetAttributes(attribute.String("payment.result_id", result.ID.String()))

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(result)
}
*/
