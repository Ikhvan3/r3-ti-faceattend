package face

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"r3-ti-faceattend/backend/internal/auth"
)

const maxFaceRequestBodyBytes = 1 << 20

type HTTPService interface {
	Status(ctx context.Context, claims auth.Claims) (StatusResponse, error)
	Enroll(ctx context.Context, claims auth.Claims, input EnrollmentInput) (StatusResponse, error)
	Reset(ctx context.Context, claims auth.Claims) error
}

type Handler struct {
	service HTTPService
}

func NewHandler(service HTTPService) Handler {
	return Handler{service: service}
}

func (h Handler) Status(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}
	status, err := h.service.Status(r.Context(), claims)
	if err != nil {
		h.writeFaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "status enrollment wajah berhasil dibaca", Data: status})
}

func (h Handler) Enroll(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}
	var req struct {
		Embedding        []float64 `json:"embedding"`
		EmbeddingModel   string    `json:"embedding_model"`
		EmbeddingVersion string    `json:"embedding_version"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	status, err := h.service.Enroll(r.Context(), claims, EnrollmentInput{
		Embedding:        req.Embedding,
		EmbeddingModel:   req.EmbeddingModel,
		EmbeddingVersion: req.EmbeddingVersion,
	})
	if err != nil {
		h.writeFaceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response{Status: "ok", Message: "enrollment wajah berhasil disimpan", Data: status})
}

func (h Handler) Enrollment(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodDelete) {
		return
	}
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}
	if err := h.service.Reset(r.Context(), claims); err != nil {
		h.writeFaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "enrollment wajah berhasil direset", Data: StatusResponse{Enrolled: false, FaceStatus: FaceStatusNotEnrolled}})
}

func (h Handler) writeFaceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrUnsupportedModel), errors.Is(err, ErrInvalidDimension):
		writeError(w, http.StatusBadRequest, "request tidak valid")
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrInactiveAccount):
		writeError(w, http.StatusForbidden, "akses enrollment wajah tidak diizinkan")
	case errors.Is(err, ErrProfileNotFound):
		writeError(w, http.StatusNotFound, "enrollment wajah tidak ditemukan")
	case errors.Is(err, ErrAlreadyEnrolled):
		writeError(w, http.StatusConflict, "wajah sudah terdaftar")
	default:
		writeError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
	}
}

type response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxFaceRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "request tidak valid")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request tidak valid")
		return false
	}
	return true
}

func allowMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
	return false
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, response{Status: "error", Message: message})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
