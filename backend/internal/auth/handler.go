package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
)

const maxRequestBodyBytes = 1 << 20

type Handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return Handler{service: service}
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	ip := clientIP(r)
	userAgent := strings.TrimSpace(r.UserAgent())
	result, err := h.service.Login(r.Context(), LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IP:        nullableString(ip),
		UserAgent: nullableString(userAgent),
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokenPayload(result))
}

func (h Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokenPayload(result))
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "logout berhasil"})
}

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "token tidak valid")
		return
	}

	profile, err := h.service.Me(r.Context(), claims)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "profil berhasil dibaca", Data: profile})
}

func AdminPing(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	writeJSON(w, http.StatusOK, response{Status: "ok", Message: "admin pong"})
}

func (h Handler) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "email atau password tidak valid")
	case errors.Is(err, ErrInvalidToken):
		writeError(w, http.StatusUnauthorized, "token tidak valid")
	case errors.Is(err, ErrInactiveAccount):
		writeError(w, http.StatusForbidden, "akun tidak aktif")
	default:
		writeError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
	}
}

type response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func tokenPayload(result TokenResponse) response {
	return response{
		Status:  "ok",
		Message: "autentikasi berhasil",
		Data: map[string]any{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"token_type":    result.TokenType,
			"expires_in":    result.ExpiresIn,
			"user":          result.User,
		},
	}
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

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}

	return host
}

func nullableString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}
