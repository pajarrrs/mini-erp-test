package handler

import (
	"net/http"
	"strings"

	"mini-erp-test/backend/internal/apperror"
	"mini-erp-test/backend/internal/model"
	"mini-erp-test/backend/internal/validation"
)

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.repo.ListOrders(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": orders})
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var in model.OrderInput
	if err := readJSON(r, &in); err != nil {
		writeError(w, err)
		return
	}
	in.Customer = strings.TrimSpace(in.Customer)
	for i := range in.Items {
		in.Items[i].SKU = strings.TrimSpace(in.Items[i].SKU)
	}

	if errs := validation.Order(in); len(errs) > 0 {
		writeError(w, apperror.Validation(errs))
		return
	}

	order, err := h.repo.CreateOrder(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	order, err := h.repo.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	order, err := h.repo.CancelOrder(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}
