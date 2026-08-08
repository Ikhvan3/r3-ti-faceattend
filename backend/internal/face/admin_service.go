package face

import (
	"context"
	"strings"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

func (s Service) AdminReset(ctx context.Context, claims auth.Claims, targetUserID string) error {
	if strings.TrimSpace(claims.Subject) == "" || claims.Role != user.RoleAdmin {
		return ErrForbidden
	}

	targetUserID = strings.TrimSpace(targetUserID)
	if targetUserID == "" {
		return ErrInvalidInput
	}

	target, err := s.users.FindByID(ctx, targetUserID)
	if err != nil || target.Role != user.RoleUser {
		return ErrProfileNotFound
	}

	if err := s.faces.DeleteByUserID(ctx, targetUserID); err != nil {
		if err == ErrProfileNotFound {
			return ErrProfileNotFound
		}
		return ErrRepositoryFailure
	}

	return nil
}
