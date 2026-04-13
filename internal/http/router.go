package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(payment *PaymentHandler, product *ProductHandler) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", payment.Health)

	r.Post("/payments", payment.Create)
	r.Get("/payments/{id}", payment.GetByID)

	r.Post("/products", product.Create)
	r.Get("/products/{id}", product.GetByID)
	r.Delete("/products/{id}", product.Delete)

	return r
}
