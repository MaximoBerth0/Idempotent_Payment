package payment

import (
	"context"
	"encoding/json"
	"idempotent-payment/internal/http/httpx"
	"log/slog"
	"net/http"
	"strings"
)

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
	Create(ctx context.Context, productID int64, idemKey string) (*Payment, error)
	GetByID(ctx context.Context, id string) (*Payment, error)
	Health(ctx context.Context) error
}

type Handler struct {
	service PaymentService
	logger  *slog.Logger
}

func NewHandler(s PaymentService, logger *slog.Logger) *Handler {
	return &Handler{service: s, logger: logger}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer r.Body.Close()

	idemKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idemKey == "" {
		h.logger.Warn("missing idempotency key", "path", r.URL.Path)
		httpx.WriteError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req CreatePaymentRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.logger.Warn("invalid JSON body", "error", err)
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.ProductID <= 0 {
		h.logger.Warn("invalid product_id", "product_id", req.ProductID)
		httpx.WriteError(w, http.StatusBadRequest, "product_id must be greater than 0")
		return
	}

	payment, err := h.service.Create(ctx, req.ProductID, idemKey)
	if err != nil {
		h.logger.Error("failed to create payment",
			"idempotency_key", idemKey,
			"error", err,
		)
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create payment")
		return
	}

	resp := CreatePaymentResponse{
		ID:        payment.ID,
		ProductID: payment.ProductID,
		Status:    string(payment.Status),
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.URL.Query().Get("id")
	if id == "" {
		h.logger.Warn(
			"missing payment ID",
			"path", r.URL.Path,
			"method", r.Method,
		)
		httpx.WriteError(w, http.StatusBadRequest, "payment ID is required")
		return
	}

	payment, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("failed to get payment", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get payment")
		return
	}

	resp := GetPaymentResponse{
		ID:        payment.ID,
		ProductID: payment.ProductID,
		Status:    string(payment.Status),
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.service.Health(ctx); err != nil {
		h.logger.Error("health check failed", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "unhealthy")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
