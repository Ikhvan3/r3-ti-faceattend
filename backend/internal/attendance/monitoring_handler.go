package attendance

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"r3-ti-faceattend/backend/internal/auth"
)

type AdminAttendanceMonitoringHTTPService interface {
	Summary(ctx context.Context, date string) (AdminAttendanceSummary, error)
	List(ctx context.Context, filter AdminAttendanceListFilter) (AdminAttendanceList, error)
	Detail(ctx context.Context, id string) (AdminAttendanceDetail, error)
}

type AdminAttendanceCorrectionHTTPService interface {
	Correct(ctx context.Context, claims auth.Claims, id string, input AdminAttendanceCorrectionInput) (AdminAttendanceDetail, error)
}

type AdminAttendanceMonitoringHandler struct {
	service    AdminAttendanceMonitoringHTTPService
	correction AdminAttendanceCorrectionHTTPService
}

func NewAdminAttendanceMonitoringHandler(service AdminAttendanceMonitoringHTTPService, correction ...AdminAttendanceCorrectionHTTPService) AdminAttendanceMonitoringHandler {
	handler := AdminAttendanceMonitoringHandler{service: service}
	if len(correction) > 0 {
		handler.correction = correction[0]
	}
	return handler
}

func (h AdminAttendanceMonitoringHandler) Summary(w http.ResponseWriter, r *http.Request) {
	if !adminAllowMethod(w, r, http.MethodGet) {
		return
	}
	result, err := h.service.Summary(r.Context(), r.URL.Query().Get("date"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "ringkasan presensi berhasil dibaca", Data: result})
}

func (h AdminAttendanceMonitoringHandler) Collection(w http.ResponseWriter, r *http.Request) {
	if !adminAllowMethod(w, r, http.MethodGet) {
		return
	}

	page, ok := parseAdminPositiveIntQuery(w, r, "page")
	if !ok {
		return
	}
	pageSize, ok := parseAdminPositiveIntQuery(w, r, "page_size")
	if !ok {
		return
	}
	isLate, ok := parseOptionalAdminBoolQuery(w, r, "is_late")
	if !ok {
		return
	}

	result, err := h.service.List(r.Context(), AdminAttendanceListFilter{
		DateFrom:        r.URL.Query().Get("date_from"),
		DateTo:          r.URL.Query().Get("date_to"),
		EmployeeID:      r.URL.Query().Get("employee_id"),
		Search:          r.URL.Query().Get("search"),
		AttendanceState: AdminAttendanceState(r.URL.Query().Get("attendance_state")),
		IsLate:          isLate,
		Page:            page,
		PageSize:        pageSize,
	})
	if err != nil {
		h.writeError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "daftar presensi berhasil dibaca", Data: result})
}

func (h AdminAttendanceMonitoringHandler) Resource(w http.ResponseWriter, r *http.Request) {
	resource := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/attendance/"), "/")
	if resource == "" {
		adminWriteError(w, http.StatusNotFound, "presensi tidak ditemukan")
		return
	}
	parts := strings.Split(resource, "/")

	if len(parts) == 2 && parts[1] == "correction" {
		h.correct(w, r, parts[0])
		return
	}
	if len(parts) != 1 {
		adminWriteError(w, http.StatusNotFound, "presensi tidak ditemukan")
		return
	}
	if !adminAllowMethod(w, r, http.MethodGet) {
		return
	}
	result, err := h.service.Detail(r.Context(), parts[0])
	if err != nil {
		h.writeError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "detail presensi berhasil dibaca", Data: result})
}

func (h AdminAttendanceMonitoringHandler) correct(w http.ResponseWriter, r *http.Request, id string) {
	if h.correction == nil {
		adminWriteError(w, http.StatusNotFound, "koreksi presensi belum tersedia")
		return
	}
	if !adminAllowMethod(w, r, http.MethodPatch) {
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		adminWriteError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	var input AdminAttendanceCorrectionInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		adminWriteError(w, http.StatusBadRequest, "data koreksi presensi tidak valid")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		adminWriteError(w, http.StatusBadRequest, "data koreksi presensi tidak valid")
		return
	}

	result, err := h.correction.Correct(r.Context(), claims, id, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrAttendanceCorrectionInvalid):
			adminWriteError(w, http.StatusBadRequest, "jam koreksi presensi tidak valid")
		case errors.Is(err, ErrAttendanceCorrectionReason):
			adminWriteError(w, http.StatusBadRequest, "alasan koreksi wajib diisi minimal 5 karakter")
		case errors.Is(err, ErrAttendanceCorrectionForbidden):
			adminWriteError(w, http.StatusForbidden, "akses admin diperlukan")
		case errors.Is(err, ErrAdminAttendanceNotFound):
			adminWriteError(w, http.StatusNotFound, "presensi tidak ditemukan")
		default:
			adminWriteError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
		}
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "koreksi presensi berhasil disimpan", Data: result})
}

func (h AdminAttendanceMonitoringHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrAdminAttendanceRange):
		adminWriteError(w, http.StatusBadRequest, "parameter presensi tidak valid")
	case errors.Is(err, ErrAdminAttendanceNotFound):
		adminWriteError(w, http.StatusNotFound, "presensi tidak ditemukan")
	default:
		adminWriteError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
	}
}

func parseOptionalAdminBoolQuery(w http.ResponseWriter, r *http.Request, key string) (*bool, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return nil, true
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		adminWriteError(w, http.StatusBadRequest, "parameter tidak valid")
		return nil, false
	}
	return &parsed, true
}
