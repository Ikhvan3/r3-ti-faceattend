package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type DatabasePinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	appEnv string
	db     DatabasePinger
}

type Response struct {
	Status      string           `json:"status"`
	Service     string           `json:"service"`
	Environment string           `json:"environment"`
	Checks      map[string]Check `json:"checks"`
}

type Check struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewHandler(appEnv string, db DatabasePinger) http.Handler {
	return &Handler{
		appEnv: appEnv,
		db:     db,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, h.response("error", Check{
			Status:  "error",
			Message: "database check was not run",
		}))
		return
	}

	databaseCheck := Check{
		Status:  "error",
		Message: "database is not connected",
	}
	statusCode := http.StatusServiceUnavailable
	status := "degraded"

	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := h.db.Ping(ctx); err == nil {
			databaseCheck = Check{
				Status:  "ok",
				Message: "database is connected",
			}
			statusCode = http.StatusOK
			status = "ok"
		}
	}

	writeJSON(w, statusCode, h.response(status, databaseCheck))
}

func (h *Handler) response(status string, databaseCheck Check) Response {
	return Response{
		Status:      status,
		Service:     "r3-ti-faceattend-api",
		Environment: h.appEnv,
		Checks: map[string]Check{
			"api": {
				Status:  "ok",
				Message: "api is running",
			},
			"database": databaseCheck,
		},
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, `{"status":"error","message":"failed to encode response"}`, http.StatusInternalServerError)
	}
}
