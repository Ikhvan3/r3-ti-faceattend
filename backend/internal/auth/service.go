package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"r3-ti-faceattend/backend/internal/config"
	"r3-ti-faceattend/backend/internal/security"
	"r3-ti-faceattend/backend/internal/user"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInactiveAccount    = errors.New("account is not active")
	ErrInternalAuth       = errors.New("authentication failed")
)

type LoginInput struct {
	Email     string
	Password  string
	IP        *string
	UserAgent *string
}

type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	User         SafeUser
}

type SafeUser struct {
	ID             string             `json:"id"`
	EmployeeNumber string             `json:"employee_number"`
	Name           string             `json:"name"`
	Email          string             `json:"email"`
	Phone          *string            `json:"phone"`
	Position       *string            `json:"position"`
	Role           user.Role          `json:"role"`
	AccountStatus  user.AccountStatus `json:"account_status"`
}

type Service struct {
	users        user.Repository
	sessions     SessionRepository
	hasher       security.PasswordHasher
	accessTokens AccessTokenService
	refresh      RefreshTokenGenerator
	refreshTTL   time.Duration
	now          func() time.Time
}

func NewService(users user.Repository, sessions SessionRepository, hasher security.PasswordHasher, cfg config.AuthConfig) Service {
	return Service{
		users:        users,
		sessions:     sessions,
		hasher:       hasher,
		accessTokens: NewAccessTokenService(cfg),
		refresh:      NewRefreshTokenGenerator(),
		refreshTTL:   cfg.RefreshTokenTTL,
		now:          time.Now,
	}
}

func (s Service) Login(ctx context.Context, input LoginInput) (TokenResponse, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" || input.Password == "" {
		return TokenResponse{}, ErrInvalidCredentials
	}

	u, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, user.ErrNotFound) {
		return TokenResponse{}, ErrInvalidCredentials
	}
	if err != nil {
		return TokenResponse{}, ErrInternalAuth
	}
	if !s.hasher.Compare(u.PasswordHash, input.Password) {
		return TokenResponse{}, ErrInvalidCredentials
	}
	if u.AccountStatus != user.AccountStatusActive {
		return TokenResponse{}, ErrInactiveAccount
	}

	return s.issue(ctx, u, "", input.IP, input.UserAgent)
}

func (s Service) Refresh(ctx context.Context, refreshToken string) (TokenResponse, error) {
	oldHash, err := HashRefreshToken(refreshToken)
	if err != nil {
		return TokenResponse{}, ErrInvalidToken
	}

	now := s.now().UTC()
	session, err := s.sessions.FindActiveByTokenHash(ctx, oldHash, now)
	if errors.Is(err, ErrSessionNotFound) {
		return TokenResponse{}, ErrInvalidToken
	}
	if err != nil {
		return TokenResponse{}, fmt.Errorf("%w: find active session", ErrInternalAuth)
	}

	u, err := s.users.FindByID(ctx, session.UserID)
	if errors.Is(err, user.ErrNotFound) {
		return TokenResponse{}, ErrInvalidToken
	}
	if err != nil {
		return TokenResponse{}, fmt.Errorf("%w: find refresh user", ErrInternalAuth)
	}
	if u.AccountStatus != user.AccountStatusActive {
		return TokenResponse{}, ErrInactiveAccount
	}

	newRefreshToken, err := s.refresh.Generate()
	if err != nil {
		return TokenResponse{}, fmt.Errorf("%w: generate refresh token", ErrInternalAuth)
	}
	newHash, err := HashRefreshToken(newRefreshToken)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("%w: hash new refresh token", ErrInternalAuth)
	}

	expiresAt := now.Add(s.refreshTTL)
	if err := s.sessions.RotateRefreshToken(ctx, session.ID, oldHash, newHash, now, expiresAt); err != nil {
		return TokenResponse{}, ErrInvalidToken
	}

	accessToken, accessExpiresAt, err := s.accessTokens.Issue(u.ID, session.ID, u.Email, u.Role)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("%w: issue access token", ErrInternalAuth)
	}

	return TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(accessExpiresAt.Sub(s.now().UTC()).Seconds()),
		User:         safeUser(u),
	}, nil
}

func (s Service) Logout(ctx context.Context, refreshToken string) error {
	hash, err := HashRefreshToken(refreshToken)
	if err != nil {
		return nil
	}

	if err := s.sessions.RevokeSession(ctx, hash, s.now().UTC()); err != nil {
		return ErrInternalAuth
	}

	return nil
}

func (s Service) Me(ctx context.Context, claims Claims) (SafeUser, error) {
	u, err := s.users.FindByID(ctx, claims.Subject)
	if errors.Is(err, user.ErrNotFound) {
		return SafeUser{}, ErrInvalidToken
	}
	if err != nil {
		return SafeUser{}, ErrInternalAuth
	}
	if u.AccountStatus != user.AccountStatusActive {
		return SafeUser{}, ErrInactiveAccount
	}

	return safeUser(u), nil
}

func (s Service) VerifyAccessToken(token string) (Claims, error) {
	return s.accessTokens.Verify(token)
}

func (s Service) issue(ctx context.Context, u user.User, sessionID string, ip *string, userAgent *string) (TokenResponse, error) {
	if sessionID == "" {
		sessionID = newUUID()
	}

	refreshToken, err := s.refresh.Generate()
	if err != nil {
		return TokenResponse{}, ErrInternalAuth
	}
	refreshHash, err := HashRefreshToken(refreshToken)
	if err != nil {
		return TokenResponse{}, ErrInternalAuth
	}

	now := s.now().UTC()
	session := Session{
		ID:               sessionID,
		UserID:           u.ID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        now.Add(s.refreshTTL),
		CreatedIP:        ip,
		UserAgent:        userAgent,
	}
	if err := s.sessions.CreateSession(ctx, session); err != nil {
		return TokenResponse{}, ErrInternalAuth
	}

	accessToken, accessExpiresAt, err := s.accessTokens.Issue(u.ID, session.ID, u.Email, u.Role)
	if err != nil {
		return TokenResponse{}, ErrInternalAuth
	}

	return TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(accessExpiresAt.Sub(s.now().UTC()).Seconds()),
		User:         safeUser(u),
	}, nil
}

func safeUser(u user.User) SafeUser {
	return SafeUser{
		ID:             u.ID,
		EmployeeNumber: u.EmployeeNumber,
		Name:           u.Name,
		Email:          u.Email,
		Phone:          u.Phone,
		Position:       u.Position,
		Role:           u.Role,
		AccountStatus:  u.AccountStatus,
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
