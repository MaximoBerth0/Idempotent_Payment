package http

import (
	"context"
	"encoding/json"
	"idempotent-payment/internal/http/httpx"
	"idempotent-payment/internal/product"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type ProductService interface {
	Create(ctx context.Context, name string, price int, currency string) (*product.Product, error)
	GetByID(ctx context.Context, id int64) (*product.Product, error)
	Delete(ctx context.Context, id int64) error
}

type ProductHandler struct {
	service ProductService
	logger  *slog.Logger
}

type CreateProductRequest struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Currency string `json:"currency"`
}

type GetProductResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Currency string `json:"currency"`
}

func NewProductHandler(s ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{service: s, logger: logger}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer r.Body.Close()

	idemKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idemKey == "" {
		h.logger.Warn("missing idempotency key", "path", r.URL.Path)
		httpx.WriteError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req CreateProductRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		h.logger.Warn("invalid JSON body", "error", err)
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		h.logger.Warn("missing product name")
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Price <= 0 {
		h.logger.Warn("invalid price", "price", req.Price)
		httpx.WriteError(w, http.StatusBadRequest, "price must be greater than 0")
		return
	}
	if strings.TrimSpace(req.Currency) == "" {
		h.logger.Warn("missing currency")
		httpx.WriteError(w, http.StatusBadRequest, "currency is required")
		return
	}

	p, err := h.service.Create(ctx, req.Name, req.Price, req.Currency)
	if err != nil {
		h.logger.Error("failed to create product",
			"name", req.Name,
			"error", err,
		)
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create product")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, p)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rawID := r.URL.Query().Get("id")
	if rawID == "" {
		h.logger.Warn(
			"missing product ID",
			"path", r.URL.Path,
			"method", r.Method,
		)
		httpx.WriteError(w, http.StatusBadRequest, "product ID is required")
		return
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		h.logger.Warn("invalid product ID", "id", rawID)
		httpx.WriteError(w, http.StatusBadRequest, "product ID must be a number")
		return
	}

	p, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("failed to get product", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get product")
		return
	}

	resp := GetProductResponse{
		ID:       p.ID,
		Name:     p.Name,
		Price:    p.Price,
		Currency: p.Currency,
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}
