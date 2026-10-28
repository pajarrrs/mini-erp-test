package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"mini-erp-test/backend/internal/apperror"
	"mini-erp-test/backend/internal/repository"
)

type Handler struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/products", h.listProducts)
	mux.HandleFunc("POST /api/products", h.createProduct)
	mux.HandleFunc("PUT /api/products/{id}", h.updateProduct)
	mux.HandleFunc("DELETE /api/products/{id}", h.deleteProduct)
	mux.HandleFunc("POST /api/products/{id}/stock", h.adjustStock)
	mux.HandleFunc("GET /api/products/{id}/movements", h.listMovements)
	mux.HandleFunc("GET /api/summary", h.getSummary)

	mux.HandleFunc("GET /api/orders", h.listOrders)
	mux.HandleFunc("POST /api/orders", h.createOrder)
	mux.HandleFunc("GET /api/orders/{id}", h.getOrder)
	mux.HandleFunc("POST /api/orders/{id}/cancel", h.cancelOrder)

	return mux
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err error) {
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		log.Printf("internal error: %v", err)
		appErr = &apperror.Error{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "Something went wrong"}
	}
	writeJSON(w, appErr.Status, map[string]any{"error": appErr})
}

func readJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			return apperror.Validation(map[string]string{typeErr.Field: "Has the wrong type (expected a whole number or text)"})
		}
		return apperror.BadRequest("Request body is not valid JSON")
	}
	return nil
}

func pathID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, apperror.NotFound("Not found")
	}
	return id, nil
}

func queryInt(r *http.Request, name string, def int) int {
	n, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil || n < 1 {
		return def
	}
	return n
}
