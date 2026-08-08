package face

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

func TestAdminResetRemovesEmployeeEnrollment(t *testing.T) {
	repo := enrolledFakeRepository([]float64{1, 0, 0})
	service := newTestService(repo)

	if err := service.AdminReset(context.Background(), testAdminClaims(), testUserID); err != nil {
		t.Fatalf("AdminReset() error = %v", err)
	}
	if _, ok := repo.profiles[testUserID]; ok {
		t.Fatal("face profile still exists after admin reset")
	}
}

func TestAdminResetRejectsNonAdmin(t *testing.T) {
	repo := enrolledFakeRepository([]float64{1, 0, 0})
	service := newTestService(repo)

	err := service.AdminReset(context.Background(), userClaims(testUserID), testUserID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("AdminReset() error = %v, want %v", err, ErrForbidden)
	}
}

func TestAdminResetRejectsAdminTarget(t *testing.T) {
	repo := enrolledFakeRepository([]float64{1, 0, 0})
	const adminTargetID = "00000000-0000-4000-8000-000000000099"
	repo.users[adminTargetID] = user.User{
		ID:            adminTargetID,
		Role:          user.RoleAdmin,
		AccountStatus: user.AccountStatusActive,
	}
	service := newTestService(repo)

	err := service.AdminReset(context.Background(), testAdminClaims(), adminTargetID)
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("AdminReset() error = %v, want %v", err, ErrProfileNotFound)
	}
}

func testAdminClaims() auth.Claims {
	return auth.Claims{
		Role: user.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "00000000-0000-4000-8000-000000000010",
		},
	}
}
