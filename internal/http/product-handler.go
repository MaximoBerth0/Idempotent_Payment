package http

import (
	"context"
	"encoding/json"
	"errors"
	"idempotent-payment/internal/http/httpx"
	"idempotent-payment/internal/product"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const productTracer = "idempotent-payment/http"

type ProductService interface {
	Create(ctx context.Context, name string, price int, currency string) (*product.Product, error)
	GetByID(ctx context.Context, id int64) (*product.Product, error)
	Delete(ctx context.Context, id int64) error
}

type ProductHandler struct {
	service ProductService
	logger  *slog.Logger
	tracer  trace.Tracer
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
	return &ProductHandler{service: s, logger: logger, tracer: otel.Tracer(productTracer)}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ProductHandler.Create")
	defer span.End()
	defer r.Body.Close()

	idemKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idemKey == "" {
		h.logger.Warn("missing idempotency key", "path", r.URL.Path)
		httpx.WriteError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	span.SetAttributes(attribute.String("payment.idempotency_key", idemKey))

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
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create product")
		h.logger.Error("failed to create product", "name", req.Name, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create product")
		return
	}

	span.SetAttributes(
		attribute.String("product.name", req.Name),
		attribute.Int("product.price", req.Price),
		attribute.String("product.currency", req.Currency),
	)
	span.SetStatus(codes.Ok, "")
	httpx.WriteJSON(w, http.StatusCreated, p)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ProductHandler.GetByID")
	defer span.End()

	rawID := chi.URLParam(r, "id")
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

	span.SetAttributes(attribute.Int64("product.id", id))

	p, err := h.service.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get product")
		h.logger.Error("failed to get product", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get product")
		return
	}

	span.SetStatus(codes.Ok, "")
	resp := GetProductResponse{
		ID:       p.ID,
		Name:     p.Name,
		Price:    p.Price,
		Currency: p.Currency,
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ProductHandler.Delete")
	defer span.End()

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid product id")
		h.logger.Error("invalid product id", "product_id", idStr, "error", err)
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int64("product.id", id))

	if err := h.service.Delete(ctx, id); err != nil {
		span.RecordError(err)
		if errors.Is(err, product.ErrNotFound) {
			span.SetStatus(codes.Error, "product not found")
			h.logger.Error("product not found", "product_id", id)
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		span.SetStatus(codes.Error, "failed to delete product")
		h.logger.Error("failed to delete product", "product_id", id, "error", err)
		http.Error(w, "failed to delete product", http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "")
	h.logger.Info("product deleted successfully", "product_id", id)
	w.WriteHeader(http.StatusNoContent)
}
