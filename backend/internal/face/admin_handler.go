package face

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"r3-ti-faceattend/backend/internal/auth"
)

type AdminHTTPService interface {
	AdminReset(ctx context.Context, claims auth.Claims, targetUserID string, reason ...string) error
}

type AdminHandler struct {
	service AdminHTTPService
}

func NewAdminHandler(service AdminHTTPService) AdminHandler {
	return AdminHandler{service: service}
}

type adminResetRequest struct {
	Reason string `json:"reason"`
}

func (h AdminHandler) ResetEnrollment(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodDelete) {
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	userID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/face-enrollments/"))
	if userID == "" || strings.Contains(userID, "/") {
		writeError(w, http.StatusBadRequest, "user pegawai tidak valid")
		return
	}

	var input adminResetRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "alasan reset enrollment wajib diisi")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request reset enrollment tidak valid")
		return
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if len(input.Reason) < 5 || len(input.Reason) > 1000 {
		writeError(w, http.StatusBadRequest, "alasan reset enrollment wajib diisi minimal 5 karakter")
		return
	}

	if err := h.service.AdminReset(r.Context(), claims, userID, input.Reason); err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "user atau alasan reset enrollment tidak valid")
		case errors.Is(err, ErrForbidden):
			writeError(w, http.StatusForbidden, "akses admin diperlukan")
		case errors.Is(err, ErrProfileNotFound):
			writeError(w, http.StatusNotFound, "enrollment wajah pegawai tidak ditemukan")
		default:
			writeError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
		}
		return
	}

	writeJSON(w, http.StatusOK, response{
		Status:  "ok",
		Message: "enrollment wajah pegawai berhasil direset",
		Data: StatusResponse{
			Enrolled:   false,
			FaceStatus: FaceStatusNotEnrolled,
		},
	})
}
