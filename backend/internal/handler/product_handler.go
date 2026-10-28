package handler

import (
	"net/http"
	"strings"

	"mini-erp-test/backend/internal/apperror"
	"mini-erp-test/backend/internal/model"
	"mini-erp-test/backend/internal/validation"
)

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := r.URL.Query().Get("status")
	page := queryInt(r, "page", 1)
	limit := min(queryInt(r, "limit", 20), 100)

	list, err := h.repo.ListProducts(r.Context(), search, status, page, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	in, err := readProductInput(r)
	if err != nil {
		writeError(w, err)
		return
	}
	product, err := h.repo.CreateProduct(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	in, err := readProductInput(r)
	if err != nil {
		writeError(w, err)
		return
	}
	product, err := h.repo.UpdateProduct(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func readProductInput(r *http.Request) (model.ProductInput, error) {
	var in model.ProductInput
	if err := readJSON(r, &in); err != nil {
		return in, err
	}
	in.SKU = strings.TrimSpace(in.SKU)
	in.Name = strings.TrimSpace(in.Name)
	if errs := validation.Product(in); len(errs) > 0 {
		return in, apperror.Validation(errs)
	}
	return in, nil
}

func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.repo.DeleteProduct(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) adjustStock(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var in model.StockInput
	if err := readJSON(r, &in); err != nil {
		writeError(w, err)
		return
	}
	in.Note = strings.TrimSpace(in.Note)
	if errs := validation.Stock(in); len(errs) > 0 {
		writeError(w, apperror.Validation(errs))
		return
	}

	product, err := h.repo.AdjustStock(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) listMovements(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	movements, err := h.repo.ListMovements(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": movements})
}

func (h *Handler) getSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.repo.GetSummary(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
