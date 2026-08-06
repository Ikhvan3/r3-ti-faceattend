package attendance

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type AdminScheduleHTTPService interface {
	ListWorkSchedules(ctx context.Context, filter WorkScheduleListFilter) (WorkScheduleList, error)
	CreateWorkSchedule(ctx context.Context, input WorkScheduleInput) (WorkSchedule, error)
	WorkScheduleDetail(ctx context.Context, id string) (WorkSchedule, error)
	UpdateWorkSchedule(ctx context.Context, id string, input WorkScheduleInput) (WorkSchedule, error)
	UpdateWorkScheduleStatus(ctx context.Context, id string, isActive bool) (WorkSchedule, error)
	ListAssignments(ctx context.Context, filter AssignmentListFilter) (ScheduleAssignmentList, error)
	CreateAssignment(ctx context.Context, input AssignmentCreateInput) (ScheduleAssignment, error)
	AssignmentDetail(ctx context.Context, id string) (ScheduleAssignment, error)
	EndAssignment(ctx context.Context, id string, effectiveTo string) (ScheduleAssignment, error)
}

type AdminScheduleHandler struct {
	service AdminScheduleHTTPService
}

func NewAdminScheduleHandler(service AdminScheduleHTTPService) AdminScheduleHandler {
	return AdminScheduleHandler{service: service}
}

func (h AdminScheduleHandler) WorkScheduleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listWorkSchedules(w, r)
	case http.MethodPost:
		h.createWorkSchedule(w, r)
	default:
		adminAllowMethod(w, r, http.MethodGet+", "+http.MethodPost)
	}
}

func (h AdminScheduleHandler) WorkScheduleResource(w http.ResponseWriter, r *http.Request) {
	id, action, ok := adminPathParts(r.URL.Path, "/api/v1/admin/work-schedules/")
	if !ok {
		adminWriteError(w, http.StatusNotFound, "jadwal kerja tidak ditemukan")
		return
	}
	if action == "status" {
		if !adminAllowMethod(w, r, http.MethodPatch) {
			return
		}
		h.updateWorkScheduleStatus(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.workScheduleDetail(w, r, id)
	case http.MethodPut:
		h.updateWorkSchedule(w, r, id)
	default:
		adminAllowMethod(w, r, http.MethodGet+", "+http.MethodPut)
	}
}

func (h AdminScheduleHandler) AssignmentCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listAssignments(w, r)
	case http.MethodPost:
		h.createAssignment(w, r)
	default:
		adminAllowMethod(w, r, http.MethodGet+", "+http.MethodPost)
	}
}

func (h AdminScheduleHandler) AssignmentResource(w http.ResponseWriter, r *http.Request) {
	id, action, ok := adminPathParts(r.URL.Path, "/api/v1/admin/schedule-assignments/")
	if !ok {
		adminWriteError(w, http.StatusNotFound, "assignment jadwal tidak ditemukan")
		return
	}
	if action == "end" {
		if !adminAllowMethod(w, r, http.MethodPatch) {
			return
		}
		h.endAssignment(w, r, id)
		return
	}

	if !adminAllowMethod(w, r, http.MethodGet) {
		return
	}
	h.assignmentDetail(w, r, id)
}

func (h AdminScheduleHandler) listWorkSchedules(w http.ResponseWriter, r *http.Request) {
	page, ok := parseAdminPositiveIntQuery(w, r, "page")
	if !ok {
		return
	}
	pageSize, ok := parseAdminPositiveIntQuery(w, r, "page_size")
	if !ok {
		return
	}
	result, err := h.service.ListWorkSchedules(r.Context(), WorkScheduleListFilter{
		Page:     page,
		PageSize: pageSize,
		Search:   r.URL.Query().Get("search"),
		Status:   ScheduleStatus(r.URL.Query().Get("status")),
	})
	if err != nil {
		h.writeAdminError(w, err)
		return
	}

	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "daftar jadwal kerja berhasil dibaca", Data: result})
}

func (h AdminScheduleHandler) createWorkSchedule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		StartTime    string `json:"start_time"`
		EndTime      string `json:"end_time"`
		GraceMinutes int    `json:"grace_minutes"`
	}
	if !adminDecodeJSON(w, r, &req) {
		return
	}
	schedule, err := h.service.CreateWorkSchedule(r.Context(), WorkScheduleInput{
		Name:         req.Name,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		GraceMinutes: req.GraceMinutes,
	})
	if err != nil {
		h.writeAdminError(w, err)
		return
	}

	adminWriteJSON(w, http.StatusCreated, adminResponse{Status: "ok", Message: "jadwal kerja berhasil dibuat", Data: schedule})
}

func (h AdminScheduleHandler) workScheduleDetail(w http.ResponseWriter, r *http.Request, id string) {
	schedule, err := h.service.WorkScheduleDetail(r.Context(), id)
	if err != nil {
		h.writeAdminError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "jadwal kerja berhasil dibaca", Data: schedule})
}

func (h AdminScheduleHandler) updateWorkSchedule(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Name         string `json:"name"`
		StartTime    string `json:"start_time"`
		EndTime      string `json:"end_time"`
		GraceMinutes int    `json:"grace_minutes"`
	}
	if !adminDecodeJSON(w, r, &req) {
		return
	}
	schedule, err := h.service.UpdateWorkSchedule(r.Context(), id, WorkScheduleInput{
		Name:         req.Name,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		GraceMinutes: req.GraceMinutes,
	})
	if err != nil {
		h.writeAdminError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "jadwal kerja berhasil diperbarui", Data: schedule})
}

func (h AdminScheduleHandler) updateWorkScheduleStatus(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		IsActive *bool `json:"is_active"`
	}
	if !adminDecodeJSON(w, r, &req) {
		return
	}
	if req.IsActive == nil {
		h.writeAdminError(w, ErrInvalidInput)
		return
	}
	schedule, err := h.service.UpdateWorkScheduleStatus(r.Context(), id, *req.IsActive)
	if err != nil {
		h.writeAdminError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "status jadwal kerja berhasil diperbarui", Data: schedule})
}

func (h AdminScheduleHandler) listAssignments(w http.ResponseWriter, r *http.Request) {
	page, ok := parseAdminPositiveIntQuery(w, r, "page")
	if !ok {
		return
	}
	pageSize, ok := parseAdminPositiveIntQuery(w, r, "page_size")
	if !ok {
		return
	}
	result, err := h.service.ListAssignments(r.Context(), AssignmentListFilter{
		Page:       page,
		PageSize:   pageSize,
		Search:     r.URL.Query().Get("search"),
		UserID:     r.URL.Query().Get("user_id"),
		ScheduleID: r.URL.Query().Get("schedule_id"),
		Status:     AssignmentStatus(r.URL.Query().Get("status")),
	})
	if err != nil {
		h.writeAdminError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "daftar assignment jadwal berhasil dibaca", Data: result})
}

func (h AdminScheduleHandler) createAssignment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID        string  `json:"user_id"`
		ScheduleID    string  `json:"schedule_id"`
		EffectiveFrom string  `json:"effective_from"`
		EffectiveTo   *string `json:"effective_to"`
	}
	if !adminDecodeJSON(w, r, &req) {
		return
	}
	assignment, err := h.service.CreateAssignment(r.Context(), AssignmentCreateInput{
		UserID:        req.UserID,
		ScheduleID:    req.ScheduleID,
		EffectiveFrom: req.EffectiveFrom,
		EffectiveTo:   req.EffectiveTo,
	})
	if err != nil {
		h.writeAdminError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusCreated, adminResponse{Status: "ok", Message: "assignment jadwal berhasil dibuat", Data: assignment})
}

func (h AdminScheduleHandler) assignmentDetail(w http.ResponseWriter, r *http.Request, id string) {
	assignment, err := h.service.AssignmentDetail(r.Context(), id)
	if err != nil {
		h.writeAdminError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "assignment jadwal berhasil dibaca", Data: assignment})
}

func (h AdminScheduleHandler) endAssignment(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		EffectiveTo string `json:"effective_to"`
	}
	if !adminDecodeJSON(w, r, &req) {
		return
	}
	assignment, err := h.service.EndAssignment(r.Context(), id, req.EffectiveTo)
	if err != nil {
		h.writeAdminError(w, err)
		return
	}
	adminWriteJSON(w, http.StatusOK, adminResponse{Status: "ok", Message: "assignment jadwal berhasil diakhiri", Data: assignment})
}

func (h AdminScheduleHandler) writeAdminError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		adminWriteError(w, http.StatusBadRequest, "request tidak valid")
	case errors.Is(err, ErrScheduleNotFound):
		adminWriteError(w, http.StatusNotFound, "jadwal kerja tidak ditemukan")
	case errors.Is(err, ErrAssignmentNotFound):
		adminWriteError(w, http.StatusNotFound, "assignment jadwal tidak ditemukan")
	case errors.Is(err, ErrScheduleDuplicate):
		adminWriteError(w, http.StatusConflict, "nama jadwal kerja sudah digunakan")
	case errors.Is(err, ErrScheduleInUse):
		adminWriteError(w, http.StatusConflict, "jadwal kerja masih memiliki assignment aktif atau masa depan")
	case errors.Is(err, ErrAssignmentOverlap):
		adminWriteError(w, http.StatusConflict, "periode assignment jadwal bertumpang tindih")
	case errors.Is(err, ErrAssignmentInvalidUser):
		adminWriteError(w, http.StatusBadRequest, "pegawai tidak valid")
	case errors.Is(err, ErrInactiveSchedule):
		adminWriteError(w, http.StatusBadRequest, "jadwal kerja tidak aktif")
	default:
		adminWriteError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
	}
}

type adminResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func adminDecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxAttendanceRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		adminWriteError(w, http.StatusBadRequest, "request tidak valid")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		adminWriteError(w, http.StatusBadRequest, "request tidak valid")
		return false
	}
	return true
}

func adminAllowMethod(w http.ResponseWriter, r *http.Request, methods string) bool {
	for _, method := range strings.Split(methods, ",") {
		if strings.TrimSpace(method) == r.Method {
			return true
		}
	}
	w.Header().Set("Allow", methods)
	adminWriteError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
	return false
}

func adminWriteError(w http.ResponseWriter, statusCode int, message string) {
	adminWriteJSON(w, statusCode, adminResponse{Status: "error", Message: message})
}

func adminWriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseAdminPositiveIntQuery(w http.ResponseWriter, r *http.Request, key string) (int, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		adminWriteError(w, http.StatusBadRequest, "parameter tidak valid")
		return 0, false
	}
	return parsed, true
}

func adminPathParts(path string, prefix string) (string, string, bool) {
	rest := strings.TrimPrefix(path, prefix)
	if rest == path || rest == "" {
		return "", "", false
	}
	parts := strings.Split(rest, "/")
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", true
	}
	if len(parts) == 2 && parts[0] != "" && (parts[1] == "status" || parts[1] == "end") {
		return parts[0], parts[1], true
	}
	return "", "", false
}
