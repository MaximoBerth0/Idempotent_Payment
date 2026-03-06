package payment

import (
	"context"
	"encoding/json"
	"net/http"
)

type CreatePaymentRequest struct {
	Amount int64 `json:"amount"`
}

type CreatePaymentResponse struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
	Status string `json:"status"`
}

type GetPaymentResponse struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
	Status string `json:"status"`
}

type PaymentService interface {
	Create(ctx context.Context, amount int64, idemKey string) (*Payment, error)
	GetByID(ctx context.Context, id string) (*Payment, error)
	Health(ctx context.Context) error
}

type Handler struct {
	service PaymentService
}

func NewHandler(s PaymentService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		http.Error(w, "Idempotency-Key header is required", http.StatusBadRequest)
		return
	}

	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	payment, err := h.service.Create(ctx, req.Amount, idemKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreatePaymentResponse{
		ID:     payment.ID,
		Amount: payment.Amount,
		Status: string(payment.Status),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	payment, err := h.service.GetByID(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := GetPaymentResponse{
		ID:     payment.ID,
		Amount: payment.Amount,
		Status: string(payment.Status),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.service.Health(ctx); err != nil {
		http.Error(w, "unhealthy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
