package location

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

const maxRequestBodyBytes = 1 << 20

type HTTPService interface {
	ListOfficeLocations(ctx context.Context, filter OfficeLocationListFilter) (OfficeLocationList, error)
	CreateOfficeLocation(ctx context.Context, input OfficeLocationInput) (OfficeLocation, error)
	OfficeLocationDetail(ctx context.Context, id string) (OfficeLocation, error)
	UpdateOfficeLocation(ctx context.Context, id string, input OfficeLocationInput) (OfficeLocation, error)
	UpdateOfficeLocationStatus(ctx context.Context, id string, isActive bool) (OfficeLocation, error)
	ListLocationAssignments(ctx context.Context, filter LocationAssignmentListFilter) (LocationAssignmentList, error)
	CreateLocationAssignment(ctx context.Context, input LocationAssignmentInput) (LocationAssignment, error)
	LocationAssignmentDetail(ctx context.Context, id string) (LocationAssignment, error)
	EndLocationAssignment(ctx context.Context, id string, effectiveTo string) (LocationAssignment, error)
	LocationRequirement(ctx context.Context, claims auth.Claims) (LocationRequirement, error)
}

type Handler struct {
	service HTTPService
}

func NewHandler(service HTTPService) Handler {
	return Handler{service: service}
}

func (h Handler) OfficeCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listOfficeLocations(w, r)
	case http.MethodPost:
		h.createOfficeLocation(w, r)
	default:
		allowMethod(w, r, http.MethodGet+", "+http.MethodPost)
	}
}

func (h Handler) OfficeResource(w http.ResponseWriter, r *http.Request) {
	id, action, ok := pathParts(r.URL.Path, "/api/v1/admin/office-locations/")
	if !ok {
		writeError(w, http.StatusNotFound, "lokasi kantor tidak ditemukan")
		return
	}
	if action == "status" {
		if !allowMethod(w, r, http.MethodPatch) {
			return
		}
		h.updateOfficeLocationStatus(w, r, id)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.officeLocationDetail(w, r, id)
	case http.MethodPut:
		h.updateOfficeLocation(w, r, id)
	default:
		allowMethod(w, r, http.MethodGet+", "+http.MethodPut)
	}
}

func (h Handler) AssignmentCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listLocationAssignments(w, r)
	case http.MethodPost:
		h.createLocationAssignment(w, r)
	default:
		allowMethod(w, r, http.MethodGet+", "+http.MethodPost)
	}
}

func (h Handler) AssignmentResource(w http.ResponseWriter, r *http.Request) {
	id, action, ok := pathParts(r.URL.Path, "/api/v1/admin/location-assignments/")
	if !ok {
		writeError(w, http.StatusNotFound, "assignment lokasi tidak ditemukan")
		return
	}
	if action == "end" {
		if !allowMethod(w, r, http.MethodPatch) {
			return
		}
		h.endLocationAssignment(w, r, id)
		return
	}
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	h.locationAssignmentDetail(w, r, id)
}

func (h Handler) LocationRequirement(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}
	requirement, err := h.service.LocationRequirement(r.Context(), claims)
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "kebutuhan lokasi absensi berhasil dibaca", Data: requirement})
}

func (h Handler) listOfficeLocations(w http.ResponseWriter, r *http.Request) {
	page, ok := parsePositiveIntQuery(w, r, "page")
	if !ok {
		return
	}
	pageSize, ok := parsePositiveIntQuery(w, r, "page_size")
	if !ok {
		return
	}
	result, err := h.service.ListOfficeLocations(r.Context(), OfficeLocationListFilter{Page: page, PageSize: pageSize, Search: r.URL.Query().Get("search"), Status: OfficeStatus(r.URL.Query().Get("status"))})
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "daftar lokasi kantor berhasil dibaca", Data: result})
}

func (h Handler) createOfficeLocation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string   `json:"name"`
		Address      *string  `json:"address"`
		Latitude     *float64 `json:"latitude"`
		Longitude    *float64 `json:"longitude"`
		RadiusMeters *int     `json:"radius_meters"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Latitude == nil || req.Longitude == nil || req.RadiusMeters == nil {
		h.writeLocationError(w, ErrInvalidInput)
		return
	}
	office, err := h.service.CreateOfficeLocation(r.Context(), OfficeLocationInput{Name: req.Name, Address: req.Address, Latitude: *req.Latitude, Longitude: *req.Longitude, RadiusMeters: *req.RadiusMeters})
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response{Status: "ok", Message: "lokasi kantor berhasil dibuat", Data: office})
}

func (h Handler) officeLocationDetail(w http.ResponseWriter, r *http.Request, id string) {
	office, err := h.service.OfficeLocationDetail(r.Context(), id)
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "lokasi kantor berhasil dibaca", Data: office})
}

func (h Handler) updateOfficeLocation(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Name         string   `json:"name"`
		Address      *string  `json:"address"`
		Latitude     *float64 `json:"latitude"`
		Longitude    *float64 `json:"longitude"`
		RadiusMeters *int     `json:"radius_meters"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Latitude == nil || req.Longitude == nil || req.RadiusMeters == nil {
		h.writeLocationError(w, ErrInvalidInput)
		return
	}
	office, err := h.service.UpdateOfficeLocation(r.Context(), id, OfficeLocationInput{Name: req.Name, Address: req.Address, Latitude: *req.Latitude, Longitude: *req.Longitude, RadiusMeters: *req.RadiusMeters})
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "lokasi kantor berhasil diperbarui", Data: office})
}

func (h Handler) updateOfficeLocationStatus(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		IsActive *bool `json:"is_active"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.IsActive == nil {
		h.writeLocationError(w, ErrInvalidInput)
		return
	}
	office, err := h.service.UpdateOfficeLocationStatus(r.Context(), id, *req.IsActive)
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "status lokasi kantor berhasil diperbarui", Data: office})
}

func (h Handler) listLocationAssignments(w http.ResponseWriter, r *http.Request) {
	page, ok := parsePositiveIntQuery(w, r, "page")
	if !ok {
		return
	}
	pageSize, ok := parsePositiveIntQuery(w, r, "page_size")
	if !ok {
		return
	}
	result, err := h.service.ListLocationAssignments(r.Context(), LocationAssignmentListFilter{
		Page: page, PageSize: pageSize, Search: r.URL.Query().Get("search"), UserID: r.URL.Query().Get("user_id"), OfficeLocationID: r.URL.Query().Get("office_location_id"), Status: AssignmentStatus(r.URL.Query().Get("status")),
	})
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "daftar assignment lokasi berhasil dibaca", Data: result})
}

func (h Handler) createLocationAssignment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID           string  `json:"user_id"`
		OfficeLocationID string  `json:"office_location_id"`
		EffectiveFrom    string  `json:"effective_from"`
		EffectiveTo      *string `json:"effective_to"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	assignment, err := h.service.CreateLocationAssignment(r.Context(), LocationAssignmentInput{UserID: req.UserID, OfficeLocationID: req.OfficeLocationID, EffectiveFrom: req.EffectiveFrom, EffectiveTo: req.EffectiveTo})
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response{Status: "ok", Message: "assignment lokasi berhasil dibuat", Data: assignment})
}

func (h Handler) locationAssignmentDetail(w http.ResponseWriter, r *http.Request, id string) {
	assignment, err := h.service.LocationAssignmentDetail(r.Context(), id)
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "assignment lokasi berhasil dibaca", Data: assignment})
}

func (h Handler) endLocationAssignment(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		EffectiveTo string `json:"effective_to"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	assignment, err := h.service.EndLocationAssignment(r.Context(), id, req.EffectiveTo)
	if err != nil {
		h.writeLocationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "assignment lokasi berhasil diakhiri", Data: assignment})
}

func (h Handler) writeLocationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "request tidak valid")
	case errors.Is(err, ErrOfficeNotFound):
		writeError(w, http.StatusNotFound, "lokasi kantor tidak ditemukan")
	case errors.Is(err, ErrAssignmentNotFound):
		writeError(w, http.StatusNotFound, "assignment lokasi tidak ditemukan")
	case errors.Is(err, ErrOfficeInUse):
		writeError(w, http.StatusConflict, "lokasi kantor masih memiliki assignment aktif atau masa depan")
	case errors.Is(err, ErrAssignmentOverlap):
		writeError(w, http.StatusConflict, "periode assignment lokasi bertumpang tindih")
	case errors.Is(err, ErrInvalidUser):
		writeError(w, http.StatusBadRequest, "pegawai tidak valid")
	case errors.Is(err, ErrInactiveOffice):
		writeError(w, http.StatusBadRequest, "lokasi kantor tidak aktif")
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
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
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

func allowMethod(w http.ResponseWriter, r *http.Request, methods string) bool {
	for _, method := range strings.Split(methods, ",") {
		if strings.TrimSpace(method) == r.Method {
			return true
		}
	}
	w.Header().Set("Allow", methods)
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

func parsePositiveIntQuery(w http.ResponseWriter, r *http.Request, key string) (int, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		writeError(w, http.StatusBadRequest, "parameter tidak valid")
		return 0, false
	}
	return parsed, true
}

func pathParts(path string, prefix string) (string, string, bool) {
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
