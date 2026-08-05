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

const maxAttendanceRequestBodyBytes = 1 << 20

type AttendanceService interface {
	Today(ctx context.Context, claims auth.Claims) (DailyStatus, error)
	CheckIn(ctx context.Context, claims auth.Claims) (DailyStatus, error)
	CheckOut(ctx context.Context, claims auth.Claims) (DailyStatus, error)
	History(ctx context.Context, claims auth.Claims, filter HistoryFilter) (HistoryList, error)
}

type Handler struct {
	service AttendanceService
}

func NewHandler(service AttendanceService) Handler {
	return Handler{service: service}
}

func (h Handler) Today(w http.ResponseWriter, r *http.Request) {
	if !attendanceAllowMethod(w, r, http.MethodGet) {
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		attendanceWriteError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	status, err := h.service.Today(r.Context(), claims)
	if err != nil {
		h.writeAttendanceError(w, err)
		return
	}

	attendanceWriteJSON(w, http.StatusOK, attendanceResponse{Status: "ok", Message: "absensi hari ini berhasil dibaca", Data: status})
}

func (h Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	if !attendanceAllowMethod(w, r, http.MethodPost) {
		return
	}
	if !emptyBody(w, r) {
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		attendanceWriteError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	status, err := h.service.CheckIn(r.Context(), claims)
	if err != nil {
		h.writeAttendanceError(w, err)
		return
	}

	attendanceWriteJSON(w, http.StatusCreated, attendanceResponse{Status: "ok", Message: "check-in berhasil", Data: status})
}

func (h Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	if !attendanceAllowMethod(w, r, http.MethodPost) {
		return
	}
	if !emptyBody(w, r) {
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		attendanceWriteError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	status, err := h.service.CheckOut(r.Context(), claims)
	if err != nil {
		h.writeAttendanceError(w, err)
		return
	}

	attendanceWriteJSON(w, http.StatusOK, attendanceResponse{Status: "ok", Message: "check-out berhasil", Data: status})
}

func (h Handler) History(w http.ResponseWriter, r *http.Request) {
	if !attendanceAllowMethod(w, r, http.MethodGet) {
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		attendanceWriteError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	page, ok := parseAttendancePositiveIntQuery(w, r, "page")
	if !ok {
		return
	}
	pageSize, ok := parseAttendancePositiveIntQuery(w, r, "page_size")
	if !ok {
		return
	}

	result, err := h.service.History(r.Context(), claims, HistoryFilter{Page: page, PageSize: pageSize})
	if err != nil {
		h.writeAttendanceError(w, err)
		return
	}

	attendanceWriteJSON(w, http.StatusOK, attendanceResponse{Status: "ok", Message: "riwayat absensi berhasil dibaca", Data: result})
}

func (h Handler) writeAttendanceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		attendanceWriteError(w, http.StatusBadRequest, "request tidak valid")
	case errors.Is(err, ErrInactiveAccount), errors.Is(err, ErrInactiveSchedule):
		attendanceWriteError(w, http.StatusForbidden, "akses absensi tidak diizinkan")
	case errors.Is(err, ErrScheduleNotFound):
		attendanceWriteError(w, http.StatusNotFound, "jadwal absensi tidak tersedia")
	case errors.Is(err, ErrAlreadyCheckedIn):
		attendanceWriteError(w, http.StatusConflict, "pegawai sudah check-in")
	case errors.Is(err, ErrNotCheckedIn):
		attendanceWriteError(w, http.StatusConflict, "pegawai belum check-in")
	case errors.Is(err, ErrAlreadyCheckedOut):
		attendanceWriteError(w, http.StatusConflict, "pegawai sudah check-out")
	default:
		attendanceWriteError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
	}
}

type attendanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func emptyBody(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxAttendanceRequestBodyBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		attendanceWriteError(w, http.StatusBadRequest, "request tidak valid")
		return false
	}
	if strings.TrimSpace(string(raw)) != "" {
		attendanceWriteError(w, http.StatusBadRequest, "request tidak valid")
		return false
	}

	return true
}

func attendanceAllowMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}

	w.Header().Set("Allow", method)
	attendanceWriteError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
	return false
}

func attendanceWriteError(w http.ResponseWriter, statusCode int, message string) {
	attendanceWriteJSON(w, statusCode, attendanceResponse{Status: "error", Message: message})
}

func attendanceWriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseAttendancePositiveIntQuery(w http.ResponseWriter, r *http.Request, key string) (int, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 0, true
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		attendanceWriteError(w, http.StatusBadRequest, "parameter tidak valid")
		return 0, false
	}

	return parsed, true
}
