package attendance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

const correctionAttendanceID = "00000000-0000-4000-8000-000000000099"

type fakeCorrectionRepository struct {
	command adminAttendanceCorrectionCommand
	err     error
}

func (f *fakeCorrectionRepository) Correct(_ context.Context, command adminAttendanceCorrectionCommand) error {
	f.command = command
	return f.err
}

type fakeCorrectionDetailReader struct {
	detail AdminAttendanceDetail
	err    error
}

func (f fakeCorrectionDetailReader) Detail(_ context.Context, _ string) (AdminAttendanceDetail, error) {
	return f.detail, f.err
}

func TestAdminAttendanceCorrectionServiceSuccess(t *testing.T) {
	repo := &fakeCorrectionRepository{}
	want := AdminAttendanceDetail{ID: correctionAttendanceID, AttendanceDate: "2026-08-08"}
	service := NewAdminAttendanceCorrectionService(
		repo,
		fakeCorrectionDetailReader{detail: want},
		time.FixedZone("Asia/Jakarta", 7*60*60),
	)
	checkOut := "17:05"

	got, err := service.Correct(context.Background(), correctionAdminClaims(), correctionAttendanceID, AdminAttendanceCorrectionInput{
		CheckInTime:  " 08:05 ",
		CheckOutTime: &checkOut,
		Reason:       "  Pegawai lupa check-out dan sudah dikonfirmasi.  ",
	})
	if err != nil {
		t.Fatalf("Correct() error = %v", err)
	}
	if got.ID != want.ID {
		t.Fatalf("detail id = %s, want %s", got.ID, want.ID)
	}
	if repo.command.CheckInTime != "08:05" || repo.command.CheckOutTime == nil || *repo.command.CheckOutTime != "17:05" {
		t.Fatalf("normalized command = %+v", repo.command)
	}
	if repo.command.Reason != "Pegawai lupa check-out dan sudah dikonfirmasi." {
		t.Fatalf("reason = %q", repo.command.Reason)
	}
	if repo.command.Actor.Subject != correctionAdminClaims().Subject {
		t.Fatalf("actor subject = %q", repo.command.Actor.Subject)
	}
}

func TestAdminAttendanceCorrectionServiceValidation(t *testing.T) {
	service := NewAdminAttendanceCorrectionService(
		&fakeCorrectionRepository{},
		fakeCorrectionDetailReader{},
		time.FixedZone("Asia/Jakarta", 7*60*60),
	)
	checkOutBefore := "07:00"

	tests := []struct {
		name   string
		claims auth.Claims
		id     string
		input  AdminAttendanceCorrectionInput
		want   error
	}{
		{
			name:   "non admin",
			claims: auth.Claims{Role: user.RoleUser, RegisteredClaims: jwt.RegisteredClaims{Subject: "00000000-0000-4000-8000-000000000010"}},
			id:     correctionAttendanceID,
			input:  AdminAttendanceCorrectionInput{CheckInTime: "08:00", Reason: "alasan valid"},
			want:   ErrAttendanceCorrectionForbidden,
		},
		{
			name:   "invalid id",
			claims: correctionAdminClaims(),
			id:     "not-a-uuid",
			input:  AdminAttendanceCorrectionInput{CheckInTime: "08:00", Reason: "alasan valid"},
			want:   ErrInvalidInput,
		},
		{
			name:   "invalid check in",
			claims: correctionAdminClaims(),
			id:     correctionAttendanceID,
			input:  AdminAttendanceCorrectionInput{CheckInTime: "25:00", Reason: "alasan valid"},
			want:   ErrAttendanceCorrectionInvalid,
		},
		{
			name:   "invalid check out format",
			claims: correctionAdminClaims(),
			id:     correctionAttendanceID,
			input:  AdminAttendanceCorrectionInput{CheckInTime: "08:00", CheckOutTime: &checkOutBefore, Reason: "alasan valid"},
			want:   nil,
		},
		{
			name:   "short reason",
			claims: correctionAdminClaims(),
			id:     correctionAttendanceID,
			input:  AdminAttendanceCorrectionInput{CheckInTime: "08:00", Reason: "abc"},
			want:   ErrAttendanceCorrectionReason,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Correct(context.Background(), tt.claims, tt.id, tt.input)
			if tt.want == nil {
				if err != nil {
					t.Fatalf("Correct() error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("Correct() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func correctionAdminClaims() auth.Claims {
	return auth.Claims{
		Email: "admin@example.test",
		Role:  user.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "00000000-0000-4000-8000-000000000001",
		},
	}
}
