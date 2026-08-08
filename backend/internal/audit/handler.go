package audit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type HTTPService interface {
	List(ctx context.Context, filter ListFilter) (List, error)
}

type Handler struct {
	service HTTPService
}

func NewHandler(service HTTPService) Handler {
	return Handler{service: service}
}

func (h Handler) Collection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	page, ok := positiveIntQuery(r, "page")
	if !ok {
		writeError(w, http.StatusBadRequest, "parameter audit log tidak valid")
		return
	}
	pageSize, ok := positiveIntQuery(r, "page_size")
	if !ok {
		writeError(w, http.StatusBadRequest, "parameter audit log tidak valid")
		return
	}

	result, err := h.service.List(r.Context(), ListFilter{
		Action:     Action(r.URL.Query().Get("action")),
		EntityType: EntityType(r.URL.Query().Get("entity_type")),
		EntityID:   r.URL.Query().Get("entity_id"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidFilter) {
			writeError(w, http.StatusBadRequest, "parameter audit log tidak valid")
			return
		}
		writeError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"message": "audit log berhasil dibaca",
		"data":    result,
	})
}

func positiveIntQuery(r *http.Request, key string) (int, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"status": "error", "message": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
