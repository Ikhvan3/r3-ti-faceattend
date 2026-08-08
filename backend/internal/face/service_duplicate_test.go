package face

import (
	"context"
	"errors"
	"testing"
	"time"
)

const secondTestUserID = "00000000-0000-4000-8000-000000000002"

func TestServiceDuplicateEnrollmentProtectionRejectsAnotherUsersBiometric(t *testing.T) {
	base := newFakeRepository()
	base.users[secondTestUserID] = base.users[testUserID]
	secondUser := base.users[secondTestUserID]
	secondUser.ID = secondTestUserID
	base.users[secondTestUserID] = secondUser

	repo := &duplicateAwareFakeRepository{fakeRepository: base}
	service := NewService(
		repo,
		repo,
		NewModelRegistry([]SupportedModel{{
			Name:             "test-face-model",
			Version:          "v1",
			Dimension:        3,
			SimilarityMetric: SimilarityMetricCosine,
			NormalizeInput:   true,
		}}),
		0.95,
		2*time.Minute,
	).WithDuplicateProtection(0.90, 20)

	if _, err := service.Enroll(context.Background(), userClaims(testUserID), EnrollmentInput{
		Embedding:        []float64{1, 0, 0},
		EmbeddingModel:   "test-face-model",
		EmbeddingVersion: "v1",
	}); err != nil {
		t.Fatalf("first Enroll() error = %v", err)
	}

	_, err := service.Enroll(context.Background(), userClaims(secondTestUserID), EnrollmentInput{
		Embedding:        []float64{0.999, 0.01, 0},
		EmbeddingModel:   "test-face-model",
		EmbeddingVersion: "v1",
	})
	if !errors.Is(err, ErrDuplicateBiometric) {
		t.Fatalf("second Enroll() error = %v, want %v", err, ErrDuplicateBiometric)
	}
	if _, ok := repo.profiles[secondTestUserID]; ok {
		t.Fatal("duplicate biometric was stored for second user")
	}
}

func TestServiceDuplicateEnrollmentProtectionAllowsDifferentBiometric(t *testing.T) {
	base := newFakeRepository()
	base.users[secondTestUserID] = base.users[testUserID]
	secondUser := base.users[secondTestUserID]
	secondUser.ID = secondTestUserID
	base.users[secondTestUserID] = secondUser

	repo := &duplicateAwareFakeRepository{fakeRepository: base}
	service := NewService(
		repo,
		repo,
		NewModelRegistry([]SupportedModel{{
			Name:             "test-face-model",
			Version:          "v1",
			Dimension:        3,
			SimilarityMetric: SimilarityMetricCosine,
			NormalizeInput:   true,
		}}),
		0.95,
		2*time.Minute,
	).WithDuplicateProtection(0.90, 20)

	if _, err := service.Enroll(context.Background(), userClaims(testUserID), EnrollmentInput{
		Embedding:        []float64{1, 0, 0},
		EmbeddingModel:   "test-face-model",
		EmbeddingVersion: "v1",
	}); err != nil {
		t.Fatalf("first Enroll() error = %v", err)
	}

	if _, err := service.Enroll(context.Background(), userClaims(secondTestUserID), EnrollmentInput{
		Embedding:        []float64{0, 1, 0},
		EmbeddingModel:   "test-face-model",
		EmbeddingVersion: "v1",
	}); err != nil {
		t.Fatalf("different biometric Enroll() error = %v", err)
	}
}

type duplicateAwareFakeRepository struct {
	*fakeRepository
}

func (r *duplicateAwareFakeRepository) CreateUnique(_ context.Context, profile FaceProfile, duplicateThreshold float64, _ int) (FaceProfile, error) {
	if _, exists := r.profiles[profile.UserID]; exists {
		return FaceProfile{}, ErrAlreadyEnrolled
	}
	for userID, existing := range r.profiles {
		if userID == profile.UserID || existing.Status != FaceStatusEnrolled {
			continue
		}
		if existing.EmbeddingModel != profile.EmbeddingModel || existing.EmbeddingVersion != profile.EmbeddingVersion {
			continue
		}
		similarity, err := CosineSimilarity(profile.Embedding, existing.Embedding)
		if err != nil {
			return FaceProfile{}, ErrRepositoryFailure
		}
		if similarity >= duplicateThreshold {
			return FaceProfile{}, ErrDuplicateBiometric
		}
	}
	return r.Create(context.Background(), profile)
}
