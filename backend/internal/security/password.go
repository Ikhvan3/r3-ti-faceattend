package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmptyPassword = errors.New("password must not be empty")

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) bool
}

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher() BcryptPasswordHasher {
	return BcryptPasswordHasher{cost: bcrypt.DefaultCost}
}

func (h BcryptPasswordHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (h BcryptPasswordHasher) Compare(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
