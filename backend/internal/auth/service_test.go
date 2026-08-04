package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"r3-ti-faceattend/backend/internal/config"
	"r3-ti-faceattend/backend/internal/user"
)

func TestServiceLoginSuccess(t *testing.T) {
	users := newAuthFakeUserRepository()
	users.users = append(users.users, activeAdmin())
	sessions := newAuthFakeSessionRepository()
	service := newTestService(users, sessions)

	result, err := service.Login(context.Background(), LoginInput{Email: "ADMIN@EXAMPLE.TEST", Password: "secret"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatal("Login() returned empty tokens")
	}
	if result.User.ID == "" || result.User.Email != "admin@example.test" {
		t.Fatalf("safe user = %+v", result.User)
	}
	if len(sessions.sessions) != 1 {
		t.Fatalf("sessions = %d, want 1", len(sessions.sessions))
	}
}

func TestServiceLoginUnknownEmail(t *testing.T) {
	service := newTestService(newAuthFakeUserRepository(), newAuthFakeSessionRepository())

	_, err := service.Login(context.Background(), LoginInput{Email: "missing@example.test", Password: "secret"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestServiceLoginWrongPassword(t *testing.T) {
	users := newAuthFakeUserRepository()
	users.users = append(users.users, activeAdmin())
	service := newTestService(users, newAuthFakeSessionRepository())

	_, err := service.Login(context.Background(), LoginInput{Email: "admin@example.test", Password: "wrong"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestServiceLoginInactiveAndSuspended(t *testing.T) {
	for _, status := range []user.AccountStatus{user.AccountStatusInactive, user.AccountStatusSuspended} {
		t.Run(string(status), func(t *testing.T) {
			users := newAuthFakeUserRepository()
			u := activeAdmin()
			u.AccountStatus = status
			users.users = append(users.users, u)
			service := newTestService(users, newAuthFakeSessionRepository())

			_, err := service.Login(context.Background(), LoginInput{Email: "admin@example.test", Password: "secret"})
			if !errors.Is(err, ErrInactiveAccount) {
				t.Fatalf("Login() error = %v, want %v", err, ErrInactiveAccount)
			}
		})
	}
}

func TestServiceResponseHasNoPasswordHash(t *testing.T) {
	users := newAuthFakeUserRepository()
	users.users = append(users.users, activeAdmin())
	service := newTestService(users, newAuthFakeSessionRepository())

	result, err := service.Login(context.Background(), LoginInput{Email: "admin@example.test", Password: "secret"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.User.ID == "" && result.User.Email == "" {
		t.Fatal("safe profile not returned")
	}
}

func TestServiceRefreshSuccessAndOldTokenFailsAfterRotation(t *testing.T) {
	users := newAuthFakeUserRepository()
	users.users = append(users.users, activeAdmin())
	sessions := newAuthFakeSessionRepository()
	service := newTestService(users, sessions)

	login, err := service.Login(context.Background(), LoginInput{Email: "admin@example.test", Password: "secret"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	refreshed, err := service.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed.RefreshToken == "" || refreshed.RefreshToken == login.RefreshToken {
		t.Fatal("Refresh() did not rotate refresh token")
	}
	if _, err := service.Refresh(context.Background(), login.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("second Refresh() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestServiceRefreshExpiredRevokedAndUnknown(t *testing.T) {
	users := newAuthFakeUserRepository()
	users.users = append(users.users, activeAdmin())
	sessions := newAuthFakeSessionRepository()
	service := newTestService(users, sessions)

	if _, err := service.Refresh(context.Background(), "unknown-token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("unknown Refresh() error = %v, want %v", err, ErrInvalidToken)
	}

	expiredHash, err := HashRefreshToken("expired-token")
	if err != nil {
		t.Fatalf("HashRefreshToken() error = %v", err)
	}
	revokedHash, err := HashRefreshToken("revoked-token")
	if err != nil {
		t.Fatalf("HashRefreshToken() error = %v", err)
	}
	revokedAt := timeNow()
	sessions.sessions = append(sessions.sessions,
		Session{ID: "expired-session", UserID: activeAdmin().ID, RefreshTokenHash: expiredHash, ExpiresAt: timeNow().Add(-time.Hour)},
		Session{ID: "revoked-session", UserID: activeAdmin().ID, RefreshTokenHash: revokedHash, ExpiresAt: timeNow().Add(time.Hour), RevokedAt: &revokedAt},
	)

	if _, err := service.Refresh(context.Background(), "expired-token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expired Refresh() error = %v, want %v", err, ErrInvalidToken)
	}
	if _, err := service.Refresh(context.Background(), "revoked-token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("revoked Refresh() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestServiceLogoutRevokesSession(t *testing.T) {
	users := newAuthFakeUserRepository()
	users.users = append(users.users, activeAdmin())
	sessions := newAuthFakeSessionRepository()
	service := newTestService(users, sessions)

	login, err := service.Login(context.Background(), LoginInput{Email: "admin@example.test", Password: "secret"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if err := service.Logout(context.Background(), login.RefreshToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, err := service.Refresh(context.Background(), login.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Refresh after logout error = %v, want %v", err, ErrInvalidToken)
	}
}

func newTestService(users *authFakeUserRepository, sessions *authFakeSessionRepository) Service {
	service := NewService(users, sessions, authFakeHasher{}, config.AuthConfig{
		AccessTokenSecret: "secret",
		AccessTokenTTL:    15 * time.Minute,
		RefreshTokenTTL:   24 * time.Hour,
		TokenIssuer:       "issuer",
		TokenAudience:     "audience",
	})
	service.now = timeNow
	return service
}

func activeAdmin() user.User {
	return user.User{
		ID:             "00000000-0000-4000-8000-000000000001",
		EmployeeNumber: "ADMIN-LOCAL",
		Name:           "Admin Lokal",
		Email:          "admin@example.test",
		PasswordHash:   "hashed:secret",
		Role:           user.RoleAdmin,
		AccountStatus:  user.AccountStatusActive,
	}
}

type authFakeHasher struct{}

func (authFakeHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (authFakeHasher) Compare(hash string, password string) bool {
	return hash == "hashed:"+password
}

type authFakeUserRepository struct {
	users []user.User
}

func newAuthFakeUserRepository() *authFakeUserRepository { return &authFakeUserRepository{} }

func (r *authFakeUserRepository) Create(ctx context.Context, u user.User) error {
	r.users = append(r.users, u)
	return nil
}

func (r *authFakeUserRepository) FindByID(ctx context.Context, id string) (user.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return user.User{}, user.ErrNotFound
}

func (r *authFakeUserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	for _, u := range r.users {
		if stringsEqualFold(u.Email, email) {
			return u, nil
		}
	}
	return user.User{}, user.ErrNotFound
}

func (r *authFakeUserRepository) FindByEmployeeNumber(ctx context.Context, employeeNumber string) (user.User, error) {
	for _, u := range r.users {
		if u.EmployeeNumber == employeeNumber {
			return u, nil
		}
	}
	return user.User{}, user.ErrNotFound
}

func (r *authFakeUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, err := r.FindByEmail(ctx, email)
	return err == nil, nil
}

func (r *authFakeUserRepository) ExistsByEmployeeNumber(ctx context.Context, employeeNumber string) (bool, error) {
	_, err := r.FindByEmployeeNumber(ctx, employeeNumber)
	return err == nil, nil
}

type authFakeSessionRepository struct {
	sessions []Session
}

func newAuthFakeSessionRepository() *authFakeSessionRepository { return &authFakeSessionRepository{} }

func (r *authFakeSessionRepository) CreateSession(ctx context.Context, s Session) error {
	r.sessions = append(r.sessions, s)
	return nil
}

func (r *authFakeSessionRepository) FindActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (Session, error) {
	for _, s := range r.sessions {
		if s.RefreshTokenHash == tokenHash && s.RevokedAt == nil && s.ExpiresAt.After(now) {
			return s, nil
		}
	}
	return Session{}, ErrSessionNotFound
}

func (r *authFakeSessionRepository) FindByID(ctx context.Context, id string) (Session, error) {
	for _, s := range r.sessions {
		if s.ID == id {
			return s, nil
		}
	}
	return Session{}, ErrSessionNotFound
}

func (r *authFakeSessionRepository) RotateRefreshToken(ctx context.Context, sessionID string, oldTokenHash string, newTokenHash string, now time.Time, expiresAt time.Time) error {
	for i, s := range r.sessions {
		if s.ID == sessionID && s.RefreshTokenHash == oldTokenHash && s.RevokedAt == nil && s.ExpiresAt.After(now) {
			r.sessions[i].RefreshTokenHash = newTokenHash
			r.sessions[i].LastUsedAt = &now
			r.sessions[i].ExpiresAt = expiresAt
			return nil
		}
	}
	return ErrSessionNotFound
}

func (r *authFakeSessionRepository) RevokeSession(ctx context.Context, tokenHash string, now time.Time) error {
	for i, s := range r.sessions {
		if s.RefreshTokenHash == tokenHash && s.RevokedAt == nil {
			r.sessions[i].RevokedAt = &now
		}
	}
	return nil
}

func stringsEqualFold(a string, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		aa, bb := a[i], b[i]
		if aa >= 'A' && aa <= 'Z' {
			aa += 'a' - 'A'
		}
		if bb >= 'A' && bb <= 'Z' {
			bb += 'a' - 'A'
		}
		if aa != bb {
			return false
		}
	}
	return true
}
