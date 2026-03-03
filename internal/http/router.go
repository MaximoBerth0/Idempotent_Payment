package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Health        http.HandlerFunc
	CreatePayment http.HandlerFunc
	GetPayment    http.HandlerFunc
}

func NewRouter(h Handlers) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.Health)
	r.Post("/payments", h.CreatePayment)
	r.Get("/payments/{id}", h.GetPayment)

	return r
}
