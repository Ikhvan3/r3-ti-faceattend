package attendance

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type AdminAttendanceMonitoringHTTPService interface {
	Summary(ctx context.Context, date string) (AdminAttendanceSummary, error)
	List(ctx context.Context, filter AdminAttendanceListFilter) (AdminAttendanceList, error)
	Detail(ctx context.Context, id string) (AdminAttendanceDetail, error)
}

type AdminAttendanceMonitoringHandler struct {
	service AdminAttendanceMonitoringHTTPService
}

func NewAdminAttendanceMonitoringHandler(service AdminAttendanceMonitoringHTTPService) AdminAttendanceMonitoringHandler {
	return AdminAttendanceMonitoringHandler{service: service}
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
	if !adminAllowMethod(w, r, http.MethodGet) {
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/attendance/")
	if id == "" || strings.Contains(id, "/") {
		adminWriteError(w, http.StatusNotFound, "presensi tidak ditemukan")
		return
	}
	result, err := h.service.Detail(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "detail presensi berhasil dibaca", Data: result})
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
