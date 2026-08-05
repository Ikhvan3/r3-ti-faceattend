package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"r3-ti-faceattend/backend/internal/config"
	"r3-ti-faceattend/backend/internal/database"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed attendance development failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	if cfg.AppEnv != "development" {
		return fmt.Errorf("APP_ENV must be development")
	}

	input, err := loadInput()
	if err != nil {
		return err
	}

	db := database.New(cfg.Database)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Pool(ctx)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	userID, err := findDevelopmentUser(ctx, pool, input.UserEmail)
	if err != nil {
		return err
	}

	scheduleID, err := upsertSchedule(ctx, pool, input)
	if err != nil {
		return err
	}

	if err := upsertAssignment(ctx, pool, userID, scheduleID, input.EffectiveFrom); err != nil {
		return err
	}

	fmt.Printf("Attendance development seed ready: user_email=%s schedule_name=%s effective_from=%s\n",
		input.UserEmail,
		input.ScheduleName,
		input.EffectiveFrom.Format("2006-01-02"),
	)

	return nil
}

type seedInput struct {
	UserEmail     string
	ScheduleName  string
	StartTime     string
	EndTime       string
	GraceMinutes  int
	EffectiveFrom time.Time
}

func loadInput() (seedInput, error) {
	graceMinutes, err := strconv.Atoi(env("DEV_ATTENDANCE_GRACE_MINUTES", "15"))
	if err != nil || graceMinutes < 0 {
		return seedInput{}, fmt.Errorf("DEV_ATTENDANCE_GRACE_MINUTES must be non-negative")
	}

	effectiveFrom := time.Now().In(time.Local)
	return seedInput{
		UserEmail:     strings.ToLower(env("DEV_ATTENDANCE_USER_EMAIL", "")),
		ScheduleName:  env("DEV_ATTENDANCE_SCHEDULE_NAME", "Jadwal Kerja Dummy TI"),
		StartTime:     env("DEV_ATTENDANCE_START_TIME", "08:00"),
		EndTime:       env("DEV_ATTENDANCE_END_TIME", "17:00"),
		GraceMinutes:  graceMinutes,
		EffectiveFrom: time.Date(effectiveFrom.Year(), effectiveFrom.Month(), effectiveFrom.Day(), 0, 0, 0, 0, time.Local),
	}, nil
}

func findDevelopmentUser(ctx context.Context, pool *pgxpool.Pool, email string) (string, error) {
	if strings.TrimSpace(email) == "" {
		return "", fmt.Errorf("DEV_ATTENDANCE_USER_EMAIL is required")
	}

	const query = `
		SELECT id
		FROM users
		WHERE lower(email) = lower($1)
			AND role = 'USER'
			AND account_status = 'ACTIVE'
	`

	var userID string
	if err := pool.QueryRow(ctx, query, email).Scan(&userID); err != nil {
		return "", fmt.Errorf("development USER account not found or not active")
	}

	return userID, nil
}

func upsertSchedule(ctx context.Context, pool *pgxpool.Pool, input seedInput) (string, error) {
	const query = `
		INSERT INTO work_schedules (
			id, name, start_time, end_time, grace_minutes, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3::time, $4::time, $5, TRUE, NOW(), NOW())
		ON CONFLICT (name) DO UPDATE
		SET start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			grace_minutes = EXCLUDED.grace_minutes,
			is_active = TRUE,
			updated_at = NOW()
		RETURNING id
	`

	var scheduleID string
	err := pool.QueryRow(ctx, query, newUUID(), input.ScheduleName, input.StartTime, input.EndTime, input.GraceMinutes).Scan(&scheduleID)
	if err != nil {
		return "", fmt.Errorf("seed work schedule: %w", err)
	}

	return scheduleID, nil
}

func upsertAssignment(ctx context.Context, pool *pgxpool.Pool, userID string, scheduleID string, effectiveFrom time.Time) error {
	const query = `
		INSERT INTO employee_schedule_assignments (
			id, user_id, schedule_id, effective_from, effective_to, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, NULL, NOW(), NOW())
		ON CONFLICT (user_id, schedule_id, effective_from) DO UPDATE
		SET effective_to = NULL,
			updated_at = NOW()
	`

	if _, err := pool.Exec(ctx, query, newUUID(), userID, scheduleID, effectiveFrom); err != nil {
		return fmt.Errorf("seed schedule assignment: %w", err)
	}

	return nil
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("generate uuid: %w", err))
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}
