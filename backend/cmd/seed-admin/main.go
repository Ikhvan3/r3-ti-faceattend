package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"r3-ti-faceattend/backend/internal/config"
	"r3-ti-faceattend/backend/internal/database"
	"r3-ti-faceattend/backend/internal/security"
	"r3-ti-faceattend/backend/internal/seed"
	"r3-ti-faceattend/backend/internal/user"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed admin failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	input, err := loadAdminInput()
	if err != nil {
		return err
	}

	cfg := config.Load()
	db := database.New(cfg.Database)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Pool(ctx)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	repo := user.NewPostgresRepository(pool)
	hasher := security.NewBcryptPasswordHasher()
	service := seed.NewAdminService(repo, hasher)

	admin, created, err := service.Seed(ctx, input)
	if err != nil {
		return err
	}

	status := "already exists"
	if created {
		status = "created"
	}

	fmt.Printf("Admin seed %s: employee_number=%s email=%s role=%s account_status=%s\n",
		status,
		admin.EmployeeNumber,
		admin.Email,
		admin.Role,
		admin.AccountStatus,
	)

	return nil
}

func loadAdminInput() (seed.AdminInput, error) {
	input := seed.AdminInput{
		EmployeeNumber: env("ADMIN_EMPLOYEE_NUMBER"),
		Name:           env("ADMIN_NAME"),
		Email:          env("ADMIN_EMAIL"),
		Password:       os.Getenv("ADMIN_PASSWORD"),
		Position:       env("ADMIN_POSITION"),
	}

	var missing []string
	if input.EmployeeNumber == "" {
		missing = append(missing, "ADMIN_EMPLOYEE_NUMBER")
	}
	if input.Name == "" {
		missing = append(missing, "ADMIN_NAME")
	}
	if input.Email == "" {
		missing = append(missing, "ADMIN_EMAIL")
	}
	if input.Password == "" {
		missing = append(missing, "ADMIN_PASSWORD")
	}
	if input.Position == "" {
		missing = append(missing, "ADMIN_POSITION")
	}
	if len(missing) > 0 {
		return seed.AdminInput{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	if len(input.Password) < seed.MinAdminPasswordLength {
		return seed.AdminInput{}, seed.ErrAdminPasswordTooShort
	}

	return input, nil
}

func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
