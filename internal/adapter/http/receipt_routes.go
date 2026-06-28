package httpapi

import "github.com/go-chi/chi/v5"

func RegisterReceiptRoutes(r chi.Router, handler *ReceiptHandler) {
	handler.RegisterRoutes(r)
}
