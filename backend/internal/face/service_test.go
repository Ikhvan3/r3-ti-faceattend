package face

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

const testUserID = "00000000-0000-4000-8000-000000000001"

func TestServiceStatus(t *testing.T) {
	t.Run("not enrolled", func(t *testing.T) {
		repo := newFakeRepository()
		service := newTestService(repo)

		status, err := service.Status(context.Background(), userClaims(testUserID))
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		if status.Enrolled || status.FaceStatus != FaceStatusNotEnrolled {
			t.Fatalf("status = %+v", status)
		}
	})

	t.Run("enrolled", func(t *testing.T) {
		repo := newFakeRepository()
		enrolledAt := time.Date(2026, 8, 7, 1, 0, 0, 0, time.UTC)
		repo.profiles[testUserID] = FaceProfile{
			ID:               "profile-id",
			UserID:           testUserID,
			Embedding:        []float64{0.1, 0.2, 0.3},
			EmbeddingModel:   "test-face-model",
			EmbeddingVersion: "v1",
			Status:           FaceStatusEnrolled,
			EnrolledAt:       &enrolledAt,
		}
		service := newTestService(repo)

		status, err := service.Status(context.Background(), userClaims(testUserID))
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		if !status.Enrolled || status.FaceStatus != FaceStatusEnrolled || status.EmbeddingModel != "test-face-model" || status.EnrolledAt == nil {
			t.Fatalf("status = %+v", status)
		}
	})
}

func TestServiceEnrollRules(t *testing.T) {
	tests := []struct {
		name  string
		input EnrollmentInput
		want  error
	}{
		{name: "empty embedding", input: EnrollmentInput{EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}, want: ErrInvalidInput},
		{name: "nan", input: EnrollmentInput{Embedding: []float64{0.1, math.NaN(), 0.3}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}, want: ErrInvalidInput},
		{name: "infinity", input: EnrollmentInput{Embedding: []float64{0.1, math.Inf(1), 0.3}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}, want: ErrInvalidInput},
		{name: "dimension mismatch", input: EnrollmentInput{Embedding: []float64{0.1, 0.2}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}, want: ErrInvalidDimension},
		{name: "unsupported model", input: EnrollmentInput{Embedding: []float64{0.1, 0.2, 0.3}, EmbeddingModel: "unknown", EmbeddingVersion: "v1"}, want: ErrUnsupportedModel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newTestService(newFakeRepository())

			_, err := service.Enroll(context.Background(), userClaims(testUserID), tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Enroll() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestServiceEnrollSuccessDuplicateAndReset(t *testing.T) {
	repo := newFakeRepository()
	service := newTestService(repo)

	status, err := service.Enroll(context.Background(), userClaims(testUserID), validInput())
	if err != nil {
		t.Fatalf("Enroll() error = %v", err)
	}
	if !status.Enrolled || status.EmbeddingModel != "test-face-model" {
		t.Fatalf("status = %+v", status)
	}
	if got := repo.profiles[testUserID].Embedding; len(got) != 3 {
		t.Fatalf("stored embedding length = %d", len(got))
	}

	_, err = service.Enroll(context.Background(), userClaims(testUserID), validInput())
	if !errors.Is(err, ErrAlreadyEnrolled) {
		t.Fatalf("duplicate Enroll() error = %v, want %v", err, ErrAlreadyEnrolled)
	}

	if err := service.Reset(context.Background(), userClaims(testUserID)); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	status, err = service.Status(context.Background(), userClaims(testUserID))
	if err != nil {
		t.Fatalf("Status() after reset error = %v", err)
	}
	if status.Enrolled || status.FaceStatus != FaceStatusNotEnrolled {
		t.Fatalf("status after reset = %+v", status)
	}

	if _, err := service.Enroll(context.Background(), userClaims(testUserID), validInput()); err != nil {
		t.Fatalf("Enroll() after reset error = %v", err)
	}
}

func TestServiceResetMissingAndOwnership(t *testing.T) {
	service := newTestService(newFakeRepository())

	if err := service.Reset(context.Background(), userClaims(testUserID)); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("Reset() error = %v, want %v", err, ErrProfileNotFound)
	}
	if _, err := service.Status(context.Background(), adminClaims()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("admin Status() error = %v, want %v", err, ErrForbidden)
	}
	if _, err := service.Enroll(context.Background(), auth.Claims{}, validInput()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("missing claims Enroll() error = %v, want %v", err, ErrForbidden)
	}
}

func TestServiceEnrollRequiresActiveUser(t *testing.T) {
	repo := newFakeRepository()
	repo.users[testUserID] = user.User{ID: testUserID, Role: user.RoleUser, AccountStatus: user.AccountStatusInactive}
	service := newTestService(repo)

	_, err := service.Enroll(context.Background(), userClaims(testUserID), validInput())
	if !errors.Is(err, ErrInactiveAccount) {
		t.Fatalf("Enroll() error = %v, want %v", err, ErrInactiveAccount)
	}
}

func newTestService(repo *fakeRepository) Service {
	service := NewService(repo, repo, NewModelRegistry([]SupportedModel{{
		Name:      "test-face-model",
		Version:   "v1",
		Dimension: 3,
	}}))
	service.now = func() time.Time { return time.Date(2026, 8, 7, 1, 0, 0, 0, time.UTC) }
	return service
}

func validInput() EnrollmentInput {
	return EnrollmentInput{Embedding: []float64{0.1, 0.2, 0.3}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}
}

func userClaims(userID string) auth.Claims {
	return auth.Claims{
		Role: user.RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
}

func adminClaims() auth.Claims {
	return auth.Claims{
		Role: user.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "admin-id",
		},
	}
}

type fakeRepository struct {
	profiles map[string]FaceProfile
	users    map[string]user.User
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		profiles: map[string]FaceProfile{},
		users: map[string]user.User{
			testUserID: {ID: testUserID, Role: user.RoleUser, AccountStatus: user.AccountStatusActive},
		},
	}
}

func (r *fakeRepository) FindByUserID(_ context.Context, userID string) (FaceProfile, error) {
	profile, ok := r.profiles[userID]
	if !ok {
		return FaceProfile{}, ErrProfileNotFound
	}
	cloned := profile
	cloned.Embedding = append([]float64(nil), profile.Embedding...)
	return cloned, nil
}

func (r *fakeRepository) Create(_ context.Context, profile FaceProfile) (FaceProfile, error) {
	if _, ok := r.profiles[profile.UserID]; ok {
		return FaceProfile{}, ErrAlreadyEnrolled
	}
	stored := profile
	stored.Embedding = append([]float64(nil), profile.Embedding...)
	stored.CreatedAt = time.Now().UTC()
	stored.UpdatedAt = stored.CreatedAt
	r.profiles[profile.UserID] = stored
	return stored, nil
}

func (r *fakeRepository) DeleteByUserID(_ context.Context, userID string) error {
	if _, ok := r.profiles[userID]; !ok {
		return ErrProfileNotFound
	}
	delete(r.profiles, userID)
	return nil
}

func (r *fakeRepository) FindByID(_ context.Context, id string) (user.User, error) {
	u, ok := r.users[id]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return u, nil
}
