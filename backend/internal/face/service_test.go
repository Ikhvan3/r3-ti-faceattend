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

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name       string
		left       []float64
		right      []float64
		want       float64
		wantErr    error
		assertNear bool
	}{
		{name: "same vector", left: []float64{1, 0, 0}, right: []float64{1, 0, 0}, want: 1, assertNear: true},
		{name: "very similar vector", left: []float64{1, 0, 0}, right: []float64{0.99, 0.01, 0}, want: 0.9999, assertNear: true},
		{name: "different vector", left: []float64{1, 0, 0}, right: []float64{0, 1, 0}, want: 0, assertNear: true},
		{name: "dimension mismatch", left: []float64{1, 0}, right: []float64{1, 0, 0}, wantErr: ErrInvalidDimension},
		{name: "zero vector", left: []float64{0, 0, 0}, right: []float64{1, 0, 0}, wantErr: ErrInvalidInput},
		{name: "nan", left: []float64{math.NaN(), 0, 0}, right: []float64{1, 0, 0}, wantErr: ErrInvalidInput},
		{name: "infinity", left: []float64{math.Inf(1), 0, 0}, right: []float64{1, 0, 0}, wantErr: ErrInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CosineSimilarity(tt.left, tt.right)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CosineSimilarity() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CosineSimilarity() error = %v", err)
			}
			if got < -1 || got > 1 {
				t.Fatalf("similarity range = %v, want [-1,1]", got)
			}
			if tt.assertNear && math.Abs(got-tt.want) > 0.001 {
				t.Fatalf("similarity = %v, want near %v", got, tt.want)
			}
			reversed, err := CosineSimilarity(tt.right, tt.left)
			if err != nil {
				t.Fatalf("reversed CosineSimilarity() error = %v", err)
			}
			if math.Abs(got-reversed) > 1e-12 {
				t.Fatalf("similarity is not symmetric: %v vs %v", got, reversed)
			}
		})
	}
}

func TestServiceVerify(t *testing.T) {
	t.Run("same embedding verifies according to threshold", func(t *testing.T) {
		repo := enrolledFakeRepository([]float64{1, 0, 0})
		service := newTestService(repo)

		result, err := service.Verify(context.Background(), userClaims(testUserID), VerificationInput{
			Embedding:        []float64{1, 0, 0},
			EmbeddingModel:   "test-face-model",
			EmbeddingVersion: "v1",
		})
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}
		if !result.Verified {
			t.Fatal("Verified = false, want true")
		}
	})

	t.Run("similar embedding verifies", func(t *testing.T) {
		repo := enrolledFakeRepository([]float64{1, 0, 0})
		service := newTestService(repo)

		result, err := service.Verify(context.Background(), userClaims(testUserID), VerificationInput{
			Embedding:        []float64{0.98, 0.05, 0},
			EmbeddingModel:   "test-face-model",
			EmbeddingVersion: "v1",
		})
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}
		if !result.Verified {
			t.Fatal("Verified = false, want true")
		}
	})

	t.Run("mismatch returns verified false", func(t *testing.T) {
		repo := enrolledFakeRepository([]float64{1, 0, 0})
		service := newTestService(repo)

		result, err := service.Verify(context.Background(), userClaims(testUserID), VerificationInput{
			Embedding:        []float64{0, 1, 0},
			EmbeddingModel:   "test-face-model",
			EmbeddingVersion: "v1",
		})
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}
		if result.Verified {
			t.Fatal("Verified = true, want false")
		}
	})

	tests := []struct {
		name  string
		repo  *fakeRepository
		input VerificationInput
		want  error
	}{
		{name: "not enrolled", repo: newFakeRepository(), input: validVerificationInput(), want: ErrNotEnrolled},
		{name: "wrong model", repo: enrolledFakeRepository([]float64{1, 0, 0}), input: VerificationInput{Embedding: []float64{1, 0, 0}, EmbeddingModel: "other", EmbeddingVersion: "v1"}, want: ErrUnsupportedModel},
		{name: "wrong version", repo: enrolledFakeRepository([]float64{1, 0, 0}), input: VerificationInput{Embedding: []float64{1, 0, 0}, EmbeddingModel: "test-face-model", EmbeddingVersion: "other"}, want: ErrUnsupportedModel},
		{name: "wrong dimension", repo: enrolledFakeRepository([]float64{1, 0, 0}), input: VerificationInput{Embedding: []float64{1, 0}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}, want: ErrInvalidDimension},
		{name: "nan", repo: enrolledFakeRepository([]float64{1, 0, 0}), input: VerificationInput{Embedding: []float64{math.NaN(), 0, 0}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}, want: ErrInvalidInput},
		{name: "infinity", repo: enrolledFakeRepository([]float64{1, 0, 0}), input: VerificationInput{Embedding: []float64{math.Inf(1), 0, 0}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}, want: ErrInvalidInput},
		{name: "zero vector", repo: enrolledFakeRepository([]float64{1, 0, 0}), input: VerificationInput{Embedding: []float64{0, 0, 0}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}, want: ErrInvalidInput},
		{name: "stored invalid", repo: enrolledFakeRepository([]float64{0, 0, 0}), input: validVerificationInput(), want: ErrRepositoryFailure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newTestService(tt.repo)
			_, err := service.Verify(context.Background(), userClaims(testUserID), tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Verify() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestServiceVerifyRequiresActiveUserAndOwnClaims(t *testing.T) {
	repo := enrolledFakeRepository([]float64{1, 0, 0})
	repo.users[testUserID] = user.User{ID: testUserID, Role: user.RoleUser, AccountStatus: user.AccountStatusInactive}
	service := newTestService(repo)

	_, err := service.Verify(context.Background(), userClaims(testUserID), validVerificationInput())
	if !errors.Is(err, ErrInactiveAccount) {
		t.Fatalf("inactive Verify() error = %v, want %v", err, ErrInactiveAccount)
	}

	repo = enrolledFakeRepository([]float64{1, 0, 0})
	service = newTestService(repo)
	_, err = service.Verify(context.Background(), adminClaims(), validVerificationInput())
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("admin Verify() error = %v, want %v", err, ErrForbidden)
	}
}

func TestProductionModelRegistryEnrollmentRules(t *testing.T) {
	t.Run("registered model accepted", func(t *testing.T) {
		service := newProductionTestService(newFakeRepository())
		_, err := service.Enroll(context.Background(), userClaims(testUserID), productionInput(FaceNetModelDimension))
		if err != nil {
			t.Fatalf("Enroll() error = %v", err)
		}
	})

	t.Run("wrong model rejected", func(t *testing.T) {
		service := newProductionTestService(newFakeRepository())
		input := productionInput(FaceNetModelDimension)
		input.EmbeddingModel = "other-model"
		_, err := service.Enroll(context.Background(), userClaims(testUserID), input)
		if !errors.Is(err, ErrUnsupportedModel) {
			t.Fatalf("Enroll() error = %v, want %v", err, ErrUnsupportedModel)
		}
	})

	t.Run("wrong version rejected", func(t *testing.T) {
		service := newProductionTestService(newFakeRepository())
		input := productionInput(FaceNetModelDimension)
		input.EmbeddingVersion = "other-version"
		_, err := service.Enroll(context.Background(), userClaims(testUserID), input)
		if !errors.Is(err, ErrUnsupportedModel) {
			t.Fatalf("Enroll() error = %v, want %v", err, ErrUnsupportedModel)
		}
	})

	t.Run("wrong dimension rejected", func(t *testing.T) {
		service := newProductionTestService(newFakeRepository())
		_, err := service.Enroll(context.Background(), userClaims(testUserID), productionInput(FaceNetModelDimension-1))
		if !errors.Is(err, ErrInvalidDimension) {
			t.Fatalf("Enroll() error = %v, want %v", err, ErrInvalidDimension)
		}
	})
}

func newTestService(repo *fakeRepository) Service {
	service := NewService(repo, repo, NewModelRegistry([]SupportedModel{{
		Name:             "test-face-model",
		Version:          "v1",
		Dimension:        3,
		SimilarityMetric: SimilarityMetricCosine,
		NormalizeInput:   true,
	}}), 0.95)
	service.now = func() time.Time { return time.Date(2026, 8, 7, 1, 0, 0, 0, time.UTC) }
	return service
}

func newProductionTestService(repo *fakeRepository) Service {
	service := NewService(repo, repo, ProductionModelRegistry(), 0.95)
	service.now = func() time.Time { return time.Date(2026, 8, 7, 1, 0, 0, 0, time.UTC) }
	return service
}

func validInput() EnrollmentInput {
	return EnrollmentInput{Embedding: []float64{0.1, 0.2, 0.3}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}
}

func validVerificationInput() VerificationInput {
	return VerificationInput{Embedding: []float64{1, 0, 0}, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1"}
}

func enrolledFakeRepository(embedding []float64) *fakeRepository {
	repo := newFakeRepository()
	enrolledAt := time.Date(2026, 8, 7, 1, 0, 0, 0, time.UTC)
	repo.profiles[testUserID] = FaceProfile{
		ID:               "profile-id",
		UserID:           testUserID,
		Embedding:        embedding,
		EmbeddingModel:   "test-face-model",
		EmbeddingVersion: "v1",
		Status:           FaceStatusEnrolled,
		EnrolledAt:       &enrolledAt,
	}
	return repo
}

func productionInput(dimension int) EnrollmentInput {
	embedding := make([]float64, dimension)
	for i := range embedding {
		embedding[i] = 1 / float64(dimension)
	}
	return EnrollmentInput{
		Embedding:        embedding,
		EmbeddingModel:   FaceNetModelName,
		EmbeddingVersion: FaceNetModelVersion,
	}
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
