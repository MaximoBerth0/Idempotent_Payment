package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"idempotent-payment/internal/http/httpx"
	"idempotent-payment/internal/payment"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const paymentTracer = "idempotent-payment/http"

type CreatePaymentRequest struct {
	ProductID int64 `json:"product_id"`
}

type CreatePaymentResponse struct {
	ID        string `json:"id"`
	ProductID int64  `json:"product_id"`
	Status    string `json:"status"`
}

type GetPaymentResponse struct {
	ID        string `json:"id"`
	ProductID int64  `json:"product_id"`
	Status    string `json:"status"`
}

type PaymentService interface {
	Create(ctx context.Context, productID int64, idemKey string) (*payment.Payment, error)
	GetByID(ctx context.Context, id string) (*payment.Payment, error)
	Health(ctx context.Context) error
}

type IdempotencyService interface {
	Execute(ctx context.Context,
		key string,
		requestHash string,
		handler func(ctx context.Context) ([]byte, int, error),
	) ([]byte, int, error)
}

type PaymentHandler struct {
	service     PaymentService
	idempotency IdempotencyService
	logger      *slog.Logger
	tracer      trace.Tracer
}

func NewPaymentHandler(s PaymentService, i IdempotencyService, logger *slog.Logger) *PaymentHandler {
	return &PaymentHandler{
		service:     s,
		idempotency: i,
		logger:      logger,
		tracer:      otel.Tracer(paymentTracer),
	}
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PaymentHandler.Create")
	defer span.End()

	defer r.Body.Close()

	idemKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idemKey == "" {
		span.SetStatus(codes.Error, "missing idempotency key")
		httpx.WriteError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	span.SetAttributes(attribute.String("payment.idempotency_key", idemKey))

	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	span.SetAttributes(attribute.Int64("payment.product_id", req.ProductID))

	requestHash := computeHash(req)

	responseBody, httpStatus, err := h.idempotency.Execute(ctx, idemKey, requestHash,
		func(ctx context.Context) ([]byte, int, error) {
			payment, err := h.service.Create(ctx, req.ProductID, idemKey)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			span.SetAttributes(attribute.String("payment.id", payment.ID))

			resp, _ := json.Marshal(CreatePaymentResponse{
				ID:        payment.ID,
				ProductID: payment.ProductID,
				Status:    string(payment.Status),
			})
			return resp, http.StatusCreated, nil
		},
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		httpx.WriteError(w, httpStatus, err.Error())
		return
	}

	span.SetAttributes(attribute.Int("http.status_code", httpStatus))
	span.SetStatus(codes.Ok, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	w.Write(responseBody)
}

func computeHash(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (h *PaymentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PaymentHandler.GetByID")
	defer span.End()

	id := chi.URLParam(r, "id")
	if id == "" {
		span.SetStatus(codes.Error, "missing payment ID")
		h.logger.Warn("missing payment ID", "path", r.URL.Path, "method", r.Method)
		httpx.WriteError(w, http.StatusBadRequest, "payment ID is required")
		return
	}

	span.SetAttributes(attribute.String("payment.id", id))

	payment, err := h.service.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get payment")
		h.logger.Error("failed to get payment", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get payment")
		return
	}

	span.SetAttributes(
		attribute.Int64("payment.product_id", payment.ProductID),
		attribute.String("payment.status", string(payment.Status)),
	)
	span.SetStatus(codes.Ok, "")

	resp := GetPaymentResponse{
		ID:        payment.ID,
		ProductID: payment.ProductID,
		Status:    string(payment.Status),
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *PaymentHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PaymentHandler.Health")
	defer span.End()

	if err := h.service.Health(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "health check failed")
		h.logger.Error("health check failed", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "unhealthy")
		return
	}

	span.SetStatus(codes.Ok, "")
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
