package security

import "testing"

func TestBcryptPasswordHasherHash(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	hash, err := hasher.Hash("admin-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "" {
		t.Fatal("Hash() returned empty hash")
	}
	if hash == "admin-password" {
		t.Fatal("Hash() returned plaintext password")
	}
}

func TestBcryptPasswordHasherCompareCorrectPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()
	hash, err := hasher.Hash("admin-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if !hasher.Compare(hash, "admin-password") {
		t.Fatal("Compare() returned false for correct password")
	}
}

func TestBcryptPasswordHasherCompareWrongPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()
	hash, err := hasher.Hash("admin-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hasher.Compare(hash, "wrong-password") {
		t.Fatal("Compare() returned true for wrong password")
	}
}

func TestBcryptPasswordHasherRejectsEmptyPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	if _, err := hasher.Hash(""); err == nil {
		t.Fatal("Hash() error = nil, want error")
	}
}
