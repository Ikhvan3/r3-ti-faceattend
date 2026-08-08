package face

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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
	CreateVerificationGrant(ctx context.Context, grant VerificationGrant) error
}

type UniqueEnrollmentRepository interface {
	CreateUnique(ctx context.Context, profile FaceProfile, duplicateThreshold float64, searchTopK int) (FaceProfile, error)
}

type UserRepository interface {
	FindByID(ctx context.Context, id string) (user.User, error)
}

type Service struct {
	faces                        Repository
	users                        UserRepository
	models                       ModelRegistry
	verificationThreshold        float64
	duplicateEnrollmentThreshold float64
	duplicateSearchTopK          int
	attendanceGrantTTL           time.Duration
	now                          func() time.Time
}

func NewService(faces Repository, users UserRepository, models ModelRegistry, verificationThreshold float64, attendanceGrantTTL time.Duration) Service {
	return Service{
		faces:                        faces,
		users:                        users,
		models:                       models,
		verificationThreshold:        verificationThreshold,
		duplicateEnrollmentThreshold: math.NaN(),
		duplicateSearchTopK:          20,
		attendanceGrantTTL:           attendanceGrantTTL,
		now:                          time.Now,
	}
}

func (s Service) WithDuplicateProtection(threshold float64, searchTopK int) Service {
	s.duplicateEnrollmentThreshold = threshold
	s.duplicateSearchTopK = searchTopK
	return s
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
	candidate := FaceProfile{
		ID:               newUUID(),
		UserID:           userID,
		Embedding:        append([]float64(nil), input.Embedding...),
		EmbeddingModel:   strings.TrimSpace(input.EmbeddingModel),
		EmbeddingVersion: strings.TrimSpace(input.EmbeddingVersion),
		Status:           FaceStatusEnrolled,
		EnrolledAt:       &enrolledAt,
	}

	var profile FaceProfile
	if duplicateProtectionEnabled(s.duplicateEnrollmentThreshold) {
		uniqueRepo, ok := s.faces.(UniqueEnrollmentRepository)
		if !ok {
			return StatusResponse{}, ErrRepositoryFailure
		}
		profile, err = uniqueRepo.CreateUnique(
			ctx,
			candidate,
			s.duplicateEnrollmentThreshold,
			s.duplicateSearchTopK,
		)
	} else {
		profile, err = s.faces.Create(ctx, candidate)
	}
	if err != nil {
		switch err {
		case ErrAlreadyEnrolled:
			return StatusResponse{}, ErrAlreadyEnrolled
		case ErrDuplicateBiometric:
			return StatusResponse{}, ErrDuplicateBiometric
		default:
			return StatusResponse{}, ErrRepositoryFailure
		}
	}

	return statusResponse(profile), nil
}

func duplicateProtectionEnabled(threshold float64) bool {
	return !math.IsNaN(threshold) && !math.IsInf(threshold, 0) && threshold >= -1 && threshold <= 1
}

func (s Service) Verify(ctx context.Context, claims auth.Claims, input VerificationInput) (VerificationResponse, error) {
	verified, err := s.verifyMatch(ctx, claims, input)
	if err != nil {
		return VerificationResponse{}, err
	}
	return VerificationResponse{Verified: verified}, nil
}

func (s Service) VerifyForAttendance(ctx context.Context, claims auth.Claims, input AttendanceVerificationInput) (AttendanceVerificationResponse, error) {
	if input.Purpose != PurposeCheckIn && input.Purpose != PurposeCheckOut {
		return AttendanceVerificationResponse{}, ErrInvalidInput
	}
	verified, err := s.verifyMatch(ctx, claims, VerificationInput{
		Embedding:        input.Embedding,
		EmbeddingModel:   input.EmbeddingModel,
		EmbeddingVersion: input.EmbeddingVersion,
	})
	if err != nil {
		return AttendanceVerificationResponse{}, err
	}
	if !verified {
		return AttendanceVerificationResponse{}, ErrVerificationMismatch
	}

	userID, err := userIDFromClaims(claims)
	if err != nil {
		return AttendanceVerificationResponse{}, err
	}
	token, tokenHash, err := newGrantToken()
	if err != nil {
		return AttendanceVerificationResponse{}, ErrRepositoryFailure
	}
	now := s.now().UTC()
	expiresAt := now.Add(s.attendanceGrantTTL)
	if err := s.faces.CreateVerificationGrant(ctx, VerificationGrant{
		ID:        newUUID(),
		UserID:    userID,
		Purpose:   input.Purpose,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}); err != nil {
		return AttendanceVerificationResponse{}, ErrRepositoryFailure
	}

	return AttendanceVerificationResponse{VerificationGrant: token, ExpiresAt: expiresAt}, nil
}

func (s Service) verifyMatch(ctx context.Context, claims auth.Claims, input VerificationInput) (bool, error) {
	userID, err := userIDFromClaims(claims)
	if err != nil {
		return false, err
	}
	if err := s.validateActiveUser(ctx, userID); err != nil {
		return false, err
	}

	profile, err := s.faces.FindByUserID(ctx, userID)
	if err != nil {
		if err == ErrProfileNotFound {
			return false, ErrNotEnrolled
		}
		return false, ErrRepositoryFailure
	}
	if profile.Status != FaceStatusEnrolled {
		return false, ErrNotEnrolled
	}

	model, err := s.validateVerificationInput(input, profile)
	if err != nil {
		return false, err
	}

	stored, err := s.validatedEmbedding(profile.Embedding, model)
	if err != nil {
		return false, ErrRepositoryFailure
	}
	candidate, err := s.validatedEmbedding(input.Embedding, model)
	if err != nil {
		return false, err
	}
	if model.NormalizeInput {
		candidate, err = L2Normalize(candidate)
		if err != nil {
			return false, err
		}
		stored, err = L2Normalize(stored)
		if err != nil {
			return false, ErrRepositoryFailure
		}
	}

	similarity, err := CosineSimilarity(candidate, stored)
	if err != nil {
		return false, err
	}
	return similarity >= s.verificationThreshold, nil
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
	_, err := s.validatedEmbedding(input.Embedding, model)
	return err
}

func (s Service) validateVerificationInput(input VerificationInput, profile FaceProfile) (SupportedModel, error) {
	modelName := strings.TrimSpace(input.EmbeddingModel)
	modelVersion := strings.TrimSpace(input.EmbeddingVersion)
	if modelName == "" || modelVersion == "" || len(input.Embedding) == 0 {
		return SupportedModel{}, ErrInvalidInput
	}
	if modelName != strings.TrimSpace(profile.EmbeddingModel) || modelVersion != strings.TrimSpace(profile.EmbeddingVersion) {
		return SupportedModel{}, ErrUnsupportedModel
	}
	model, ok := s.models.Find(modelName, modelVersion)
	if !ok || model.Dimension < 1 {
		return SupportedModel{}, ErrUnsupportedModel
	}
	if model.SimilarityMetric != SimilarityMetricCosine {
		return SupportedModel{}, ErrUnsupportedModel
	}
	if len(input.Embedding) != model.Dimension {
		return SupportedModel{}, ErrInvalidDimension
	}
	return model, nil
}

func (s Service) validatedEmbedding(embedding []float64, model SupportedModel) ([]float64, error) {
	if len(embedding) == 0 {
		return nil, ErrInvalidInput
	}
	if len(embedding) != model.Dimension {
		return nil, ErrInvalidDimension
	}
	for _, value := range embedding {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, ErrInvalidInput
		}
	}
	return append([]float64(nil), embedding...), nil
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

func newGrantToken() (string, string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b[:])
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}
