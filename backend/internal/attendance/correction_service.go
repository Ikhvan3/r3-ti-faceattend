package attendance

import (
	"context"
	"strings"
	"time"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

type AdminAttendanceCorrectionRepository interface {
	Correct(ctx context.Context, command adminAttendanceCorrectionCommand) error
}

type AdminAttendanceDetailReader interface {
	Detail(ctx context.Context, id string) (AdminAttendanceDetail, error)
}

type AdminAttendanceCorrectionService struct {
	repo     AdminAttendanceCorrectionRepository
	detail   AdminAttendanceDetailReader
	location *time.Location
}

func NewAdminAttendanceCorrectionService(repo AdminAttendanceCorrectionRepository, detail AdminAttendanceDetailReader, location *time.Location) AdminAttendanceCorrectionService {
	return AdminAttendanceCorrectionService{repo: repo, detail: detail, location: location}
}

func (s AdminAttendanceCorrectionService) Correct(ctx context.Context, claims auth.Claims, id string, input AdminAttendanceCorrectionInput) (AdminAttendanceDetail, error) {
	if strings.TrimSpace(claims.Subject) == "" || claims.Role != user.RoleAdmin {
		return AdminAttendanceDetail{}, ErrForbidden
	}

	id = strings.TrimSpace(id)
	if !validAdminUUID(id) {
		return AdminAttendanceDetail{}, ErrInvalidInput
	}

	checkIn := strings.TrimSpace(input.CheckInTime)
	if _, err := time.Parse("15:04", checkIn); err != nil {
		return AdminAttendanceDetail{}, ErrAttendanceCorrectionInvalid
	}

	var checkOut *string
	if input.CheckOutTime != nil {
		value := strings.TrimSpace(*input.CheckOutTime)
		if value != "" {
			if _, err := time.Parse("15:04", value); err != nil {
				return AdminAttendanceDetail{}, ErrAttendanceCorrectionInvalid
			}
			checkOut = &value
		}
	}

	reason := strings.TrimSpace(input.Reason)
	if len(reason) < 5 || len(reason) > 1000 {
		return AdminAttendanceDetail{}, ErrAttendanceCorrectionReason
	}

	if err := s.repo.Correct(ctx, adminAttendanceCorrectionCommand{
		AttendanceID: id,
		CheckInTime:  checkIn,
		CheckOutTime: checkOut,
		Reason:       reason,
		Timezone:     s.location.String(),
		Actor:        claims,
	}); err != nil {
		return AdminAttendanceDetail{}, err
	}

	return s.detail.Detail(ctx, id)
}
