package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/config"
	"r3-ti-faceattend/backend/internal/user"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	SessionID string    `json:"session_id"`
	Email     string    `json:"email"`
	Role      user.Role `json:"role"`
	jwt.RegisteredClaims
}

type AccessTokenService struct {
	secret   []byte
	ttl      time.Duration
	issuer   string
	audience string
	now      func() time.Time
}

func NewAccessTokenService(cfg config.AuthConfig) AccessTokenService {
	return AccessTokenService{
		secret:   []byte(cfg.AccessTokenSecret),
		ttl:      cfg.AccessTokenTTL,
		issuer:   cfg.TokenIssuer,
		audience: cfg.TokenAudience,
		now:      time.Now,
	}
}

func (s AccessTokenService) Issue(userID string, sessionID string, email string, role user.Role) (string, time.Time, error) {
	now := s.now().UTC()
	expiresAt := now.Add(s.ttl)
	claims := Claims{
		SessionID: sessionID,
		Email:     email,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (s AccessTokenService) Verify(rawToken string) (Claims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}

		return s.secret, nil
	}, jwt.WithIssuer(s.issuer), jwt.WithAudience(s.audience), jwt.WithExpirationRequired(), jwt.WithTimeFunc(s.now))
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, ErrInvalidToken
	}
	if claims.Subject == "" || claims.SessionID == "" || claims.Email == "" || claims.Role == "" {
		return Claims{}, fmt.Errorf("%w: missing claims", ErrInvalidToken)
	}

	return *claims, nil
}

func (s AccessTokenService) ExpiresInSeconds() int {
	return int(s.ttl.Seconds())
}
