package face

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"strings"
	"time"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

type Repository interface {
	FindByUserID(ctx context.Context, userID string) (FaceProfile, error)
	Create(ctx context.Context, profile FaceProfile) (FaceProfile, error)
	DeleteByUserID(ctx context.Context, userID string) error
}

type UserRepository interface {
	FindByID(ctx context.Context, id string) (user.User, error)
}

type Service struct {
	faces  Repository
	users  UserRepository
	models ModelRegistry
	now    func() time.Time
}

func NewService(faces Repository, users UserRepository, models ModelRegistry) Service {
	return Service{faces: faces, users: users, models: models, now: time.Now}
}

func (s Service) Status(ctx context.Context, claims auth.Claims) (StatusResponse, error) {
	userID, err := userIDFromClaims(claims)
	if err != nil {
		return StatusResponse{}, err
	}

	profile, err := s.faces.FindByUserID(ctx, userID)
	if err != nil {
		if err == ErrProfileNotFound {
			return StatusResponse{Enrolled: false, FaceStatus: FaceStatusNotEnrolled}, nil
		}
		return StatusResponse{}, ErrRepositoryFailure
	}

	return statusResponse(profile), nil
}

func (s Service) Enroll(ctx context.Context, claims auth.Claims, input EnrollmentInput) (StatusResponse, error) {
	userID, err := userIDFromClaims(claims)
	if err != nil {
		return StatusResponse{}, err
	}
	if err := s.validateActiveUser(ctx, userID); err != nil {
		return StatusResponse{}, err
	}
	if err := s.validateEnrollmentInput(input); err != nil {
		return StatusResponse{}, err
	}

	existing, err := s.faces.FindByUserID(ctx, userID)
	if err != nil && err != ErrProfileNotFound {
		return StatusResponse{}, ErrRepositoryFailure
	}
	if err == nil && existing.Status == FaceStatusEnrolled {
		return StatusResponse{}, ErrAlreadyEnrolled
	}

	enrolledAt := s.now().UTC()
	profile, err := s.faces.Create(ctx, FaceProfile{
		ID:               newUUID(),
		UserID:           userID,
		Embedding:        append([]float64(nil), input.Embedding...),
		EmbeddingModel:   strings.TrimSpace(input.EmbeddingModel),
		EmbeddingVersion: strings.TrimSpace(input.EmbeddingVersion),
		Status:           FaceStatusEnrolled,
		EnrolledAt:       &enrolledAt,
	})
	if err != nil {
		if err == ErrAlreadyEnrolled {
			return StatusResponse{}, ErrAlreadyEnrolled
		}
		return StatusResponse{}, ErrRepositoryFailure
	}

	return statusResponse(profile), nil
}

func (s Service) Reset(ctx context.Context, claims auth.Claims) error {
	userID, err := userIDFromClaims(claims)
	if err != nil {
		return err
	}

	if err := s.faces.DeleteByUserID(ctx, userID); err != nil {
		if err == ErrProfileNotFound {
			return ErrProfileNotFound
		}
		return ErrRepositoryFailure
	}
	return nil
}

func (s Service) validateActiveUser(ctx context.Context, userID string) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return ErrForbidden
	}
	if u.Role != user.RoleUser {
		return ErrForbidden
	}
	if u.AccountStatus != user.AccountStatusActive {
		return ErrInactiveAccount
	}
	return nil
}

func (s Service) validateEnrollmentInput(input EnrollmentInput) error {
	modelName := strings.TrimSpace(input.EmbeddingModel)
	modelVersion := strings.TrimSpace(input.EmbeddingVersion)
	if modelName == "" || modelVersion == "" || len(input.Embedding) == 0 {
		return ErrInvalidInput
	}
	model, ok := s.models.Find(modelName, modelVersion)
	if !ok || model.Dimension < 1 {
		return ErrUnsupportedModel
	}
	if len(input.Embedding) != model.Dimension {
		return ErrInvalidDimension
	}
	for _, value := range input.Embedding {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return ErrInvalidInput
		}
	}
	return nil
}

func userIDFromClaims(claims auth.Claims) (string, error) {
	if claims.Subject == "" || claims.Role != user.RoleUser {
		return "", ErrForbidden
	}
	return claims.Subject, nil
}

func statusResponse(profile FaceProfile) StatusResponse {
	return StatusResponse{
		Enrolled:         profile.Status == FaceStatusEnrolled,
		FaceStatus:       profile.Status,
		EmbeddingModel:   profile.EmbeddingModel,
		EmbeddingVersion: profile.EmbeddingVersion,
		EnrolledAt:       profile.EnrolledAt,
	}
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
