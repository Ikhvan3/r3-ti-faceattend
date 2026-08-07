package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"r3-ti-faceattend/backend/internal/attendance"
	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/config"
	"r3-ti-faceattend/backend/internal/database"
	"r3-ti-faceattend/backend/internal/face"
	"r3-ti-faceattend/backend/internal/health"
	officelocation "r3-ti-faceattend/backend/internal/location"
	"r3-ti-faceattend/backend/internal/security"
	"r3-ti-faceattend/backend/internal/user"
)

func Run() {
	if err := run(); err != nil {
		log.Fatalf("api stopped: %v", err)
	}
}

func run() error {
	cfg := config.Load()
	if err := cfg.Auth.Validate(); err != nil {
		return err
	}
	if _, err := cfg.BusinessLocation(); err != nil {
		return err
	}
	if err := cfg.Geofence.Validate(); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := database.New(cfg.Database)
	if err := db.Ping(ctx); err != nil {
		log.Print("database is not ready")
	}
	defer db.Close()

	handler := newHTTPHandler(cfg, db)
	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("api listening on %s", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	return waitForServer(ctx, server, errCh)
}

func waitForServer(ctx context.Context, server *http.Server, errCh <-chan error) error {
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func newHTTPHandler(cfg config.Config, db health.DatabasePinger) http.Handler {
	mux := http.NewServeMux()
	healthHandler := health.NewHandler(cfg.AppEnv, db)
	mux.Handle("/health", healthHandler)
	mux.Handle("/api/v1/health", healthHandler)

	client, ok := db.(*database.Client)
	if ok && cfg.Auth.Validate() == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pool, err := client.Pool(ctx)
		if err == nil {
			userRepo := user.NewPostgresRepository(pool)
			sessionRepo := auth.NewPostgresSessionRepository(pool)
			hasher := security.NewBcryptPasswordHasher()
			authService := auth.NewService(userRepo, sessionRepo, hasher, cfg.Auth)
			authHandler := auth.NewHandler(authService)
			businessLocation, err := cfg.BusinessLocation()
			if err != nil {
				return mux
			}
			employeeService := user.NewEmployeeService(userRepo, hasher)
			employeeHandler := user.NewEmployeeHandler(employeeService)
			attendanceRepo := attendance.NewPostgresRepository(pool)
			attendanceService := attendance.NewService(attendanceRepo, businessLocation, cfg.Geofence.MaxAccuracyMeters)
			attendanceHandler := attendance.NewHandler(attendanceService)
			adminScheduleRepo := attendance.NewAdminPostgresRepository(pool)
			adminScheduleService := attendance.NewAdminScheduleService(adminScheduleRepo, businessLocation)
			adminScheduleHandler := attendance.NewAdminScheduleHandler(adminScheduleService)
			locationRepo := officelocation.NewPostgresRepository(pool)
			locationService := officelocation.NewService(locationRepo, locationRepo, businessLocation)
			locationHandler := officelocation.NewHandler(locationService)
			faceRepo := face.NewPostgresRepository(pool)
			faceService := face.NewService(faceRepo, userRepo, face.EmptyModelRegistry())
			faceHandler := face.NewHandler(faceService)
			adminOnly := func(next http.Handler) http.Handler {
				return auth.Authenticate(authService, auth.RequireRole(user.RoleAdmin, next))
			}
			userOnly := func(next http.Handler) http.Handler {
				return auth.Authenticate(authService, auth.RequireRole(user.RoleUser, next))
			}

			mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
			mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
			mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)
			mux.Handle("/api/v1/auth/me", auth.Authenticate(authService, http.HandlerFunc(authHandler.Me)))
			mux.Handle("/api/v1/admin/ping", adminOnly(http.HandlerFunc(auth.AdminPing)))
			mux.Handle("/api/v1/admin/employees", adminOnly(http.HandlerFunc(employeeHandler.Collection)))
			mux.Handle("/api/v1/admin/employees/", adminOnly(http.HandlerFunc(employeeHandler.Resource)))
			mux.Handle("/api/v1/admin/work-schedules", adminOnly(http.HandlerFunc(adminScheduleHandler.WorkScheduleCollection)))
			mux.Handle("/api/v1/admin/work-schedules/", adminOnly(http.HandlerFunc(adminScheduleHandler.WorkScheduleResource)))
			mux.Handle("/api/v1/admin/schedule-assignments", adminOnly(http.HandlerFunc(adminScheduleHandler.AssignmentCollection)))
			mux.Handle("/api/v1/admin/schedule-assignments/", adminOnly(http.HandlerFunc(adminScheduleHandler.AssignmentResource)))
			mux.Handle("/api/v1/admin/office-locations", adminOnly(http.HandlerFunc(locationHandler.OfficeCollection)))
			mux.Handle("/api/v1/admin/office-locations/", adminOnly(http.HandlerFunc(locationHandler.OfficeResource)))
			mux.Handle("/api/v1/admin/location-assignments", adminOnly(http.HandlerFunc(locationHandler.AssignmentCollection)))
			mux.Handle("/api/v1/admin/location-assignments/", adminOnly(http.HandlerFunc(locationHandler.AssignmentResource)))
			mux.Handle("/api/v1/attendance/today", userOnly(http.HandlerFunc(attendanceHandler.Today)))
			mux.Handle("/api/v1/attendance/location-requirement", userOnly(http.HandlerFunc(locationHandler.LocationRequirement)))
			mux.Handle("/api/v1/attendance/check-in", userOnly(http.HandlerFunc(attendanceHandler.CheckIn)))
			mux.Handle("/api/v1/attendance/check-out", userOnly(http.HandlerFunc(attendanceHandler.CheckOut)))
			mux.Handle("/api/v1/attendance/history", userOnly(http.HandlerFunc(attendanceHandler.History)))
			mux.Handle("/api/v1/face/status", userOnly(http.HandlerFunc(faceHandler.Status)))
			mux.Handle("/api/v1/face/enroll", userOnly(http.HandlerFunc(faceHandler.Enroll)))
			mux.Handle("/api/v1/face/enrollment", userOnly(http.HandlerFunc(faceHandler.Enrollment)))
		}
	}

	return mux
}
