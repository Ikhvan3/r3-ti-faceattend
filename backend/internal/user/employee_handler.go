package user

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const maxEmployeeRequestBodyBytes = 1 << 20

type EmployeeHandler struct {
	service EmployeeService
}

func NewEmployeeHandler(service EmployeeService) EmployeeHandler {
	return EmployeeHandler{service: service}
}

func (h EmployeeHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		employeeAllowMethod(w, r, http.MethodGet+", "+http.MethodPost)
	}
}

func (h EmployeeHandler) Resource(w http.ResponseWriter, r *http.Request) {
	id, action, ok := employeePathParts(r.URL.Path)
	if !ok {
		employeeWriteError(w, http.StatusNotFound, "pegawai tidak ditemukan")
		return
	}

	if action == "status" {
		if !employeeAllowMethod(w, r, http.MethodPatch) {
			return
		}
		h.updateStatus(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.detail(w, r, id)
	case http.MethodPut:
		h.update(w, r, id)
	default:
		employeeAllowMethod(w, r, http.MethodGet+", "+http.MethodPut)
	}
}

func (h EmployeeHandler) list(w http.ResponseWriter, r *http.Request) {
	page, ok := parsePositiveIntQuery(w, r, "page")
	if !ok {
		return
	}
	pageSize, ok := parsePositiveIntQuery(w, r, "page_size")
	if !ok {
		return
	}

	result, err := h.service.List(r.Context(), EmployeeListFilter{
		Page:     page,
		PageSize: pageSize,
		Search:   r.URL.Query().Get("search"),
		Status:   AccountStatus(r.URL.Query().Get("status")),
	})
	if err != nil {
		h.writeEmployeeError(w, err)
		return
	}

	employeeWriteJSON(w, http.StatusOK, employeeResponse{Status: "ok", Message: "daftar pegawai berhasil dibaca", Data: result})
}

func (h EmployeeHandler) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EmployeeNumber  string  `json:"employee_number"`
		Name            string  `json:"name"`
		Email           string  `json:"email"`
		InitialPassword string  `json:"initial_password"`
		Phone           *string `json:"phone"`
		Position        *string `json:"position"`
	}
	if !employeeDecodeJSON(w, r, &req) {
		return
	}

	profile, err := h.service.Create(r.Context(), EmployeeCreateInput{
		EmployeeNumber:  req.EmployeeNumber,
		Name:            req.Name,
		Email:           req.Email,
		InitialPassword: req.InitialPassword,
		Phone:           req.Phone,
		Position:        req.Position,
	})
	if err != nil {
		h.writeEmployeeError(w, err)
		return
	}

	employeeWriteJSON(w, http.StatusCreated, employeeResponse{Status: "ok", Message: "pegawai berhasil dibuat", Data: profile})
}

func (h EmployeeHandler) detail(w http.ResponseWriter, r *http.Request, id string) {
	profile, err := h.service.Detail(r.Context(), id)
	if err != nil {
		h.writeEmployeeError(w, err)
		return
	}

	employeeWriteJSON(w, http.StatusOK, employeeResponse{Status: "ok", Message: "pegawai berhasil dibaca", Data: profile})
}

func (h EmployeeHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		EmployeeNumber string  `json:"employee_number"`
		Name           string  `json:"name"`
		Email          string  `json:"email"`
		Phone          *string `json:"phone"`
		Position       *string `json:"position"`
	}
	if !employeeDecodeJSON(w, r, &req) {
		return
	}

	profile, err := h.service.Update(r.Context(), id, EmployeeUpdateInput{
		EmployeeNumber: req.EmployeeNumber,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		Position:       req.Position,
	})
	if err != nil {
		h.writeEmployeeError(w, err)
		return
	}

	employeeWriteJSON(w, http.StatusOK, employeeResponse{Status: "ok", Message: "pegawai berhasil diperbarui", Data: profile})
}

func (h EmployeeHandler) updateStatus(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		AccountStatus AccountStatus `json:"account_status"`
	}
	if !employeeDecodeJSON(w, r, &req) {
		return
	}

	profile, err := h.service.UpdateStatus(r.Context(), id, req.AccountStatus)
	if err != nil {
		h.writeEmployeeError(w, err)
		return
	}

	employeeWriteJSON(w, http.StatusOK, employeeResponse{Status: "ok", Message: "status pegawai berhasil diperbarui", Data: profile})
}

func (h EmployeeHandler) writeEmployeeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidEmployeeInput), errors.Is(err, ErrEmployeePasswordShort), errors.Is(err, ErrInvalidEmployeeStatus):
		employeeWriteError(w, http.StatusBadRequest, "request tidak valid")
	case errors.Is(err, ErrNotFound):
		employeeWriteError(w, http.StatusNotFound, "pegawai tidak ditemukan")
	case errors.Is(err, ErrEmailConflict):
		employeeWriteError(w, http.StatusConflict, "email sudah digunakan")
	case errors.Is(err, ErrEmployeeNumberConflict):
		employeeWriteError(w, http.StatusConflict, "nomor pegawai sudah digunakan")
	default:
		employeeWriteError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
	}
}

type employeeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func employeeDecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxEmployeeRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		employeeWriteError(w, http.StatusBadRequest, "request tidak valid")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		employeeWriteError(w, http.StatusBadRequest, "request tidak valid")
		return false
	}

	return true
}

func employeeAllowMethod(w http.ResponseWriter, r *http.Request, methods string) bool {
	for _, method := range strings.Split(methods, ",") {
		if strings.TrimSpace(method) == r.Method {
			return true
		}
	}

	w.Header().Set("Allow", methods)
	employeeWriteError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
	return false
}

func employeeWriteError(w http.ResponseWriter, statusCode int, message string) {
	employeeWriteJSON(w, statusCode, employeeResponse{Status: "error", Message: message})
}

func employeeWriteJSON(w http.ResponseWriter, statusCode int, payload any) {
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
		employeeWriteError(w, http.StatusBadRequest, "parameter tidak valid")
		return 0, false
	}

	return parsed, true
}

func employeePathParts(path string) (string, string, bool) {
	const prefix = "/api/v1/admin/employees/"
	rest := strings.TrimPrefix(path, prefix)
	if rest == path || rest == "" {
		return "", "", false
	}

	parts := strings.Split(rest, "/")
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", true
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "status" {
		return parts[0], "status", true
	}

	return "", "", false
}
