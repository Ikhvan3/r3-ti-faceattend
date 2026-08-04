package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

var ErrEmptyRefreshToken = errors.New("refresh token must not be empty")

type RefreshTokenGenerator struct {
	byteLength int
}

func NewRefreshTokenGenerator() RefreshTokenGenerator {
	return RefreshTokenGenerator{byteLength: 32}
}

func (g RefreshTokenGenerator) Generate() (string, error) {
	bytes := make([]byte, g.byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashRefreshToken(token string) (string, error) {
	if token == "" {
		return "", ErrEmptyRefreshToken
	}

	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:]), nil
}

func CompareRefreshTokenHash(hash string, token string) bool {
	tokenHash, err := HashRefreshToken(token)
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(hash), []byte(tokenHash)) == 1
}
