package auth

import "testing"

func TestRefreshTokenGeneratorCreatesToken(t *testing.T) {
	token, err := NewRefreshTokenGenerator().Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if token == "" {
		t.Fatal("Generate() returned empty token")
	}
}

func TestRefreshTokenHashAndCompare(t *testing.T) {
	hash, err := HashRefreshToken("refresh-token")
	if err != nil {
		t.Fatalf("HashRefreshToken() error = %v", err)
	}
	if hash == "refresh-token" {
		t.Fatal("hash equals plaintext token")
	}
	if !CompareRefreshTokenHash(hash, "refresh-token") {
		t.Fatal("CompareRefreshTokenHash() returned false for correct token")
	}
	if CompareRefreshTokenHash(hash, "wrong-token") {
		t.Fatal("CompareRefreshTokenHash() returned true for wrong token")
	}
}

func TestRefreshTokenHashRejectsEmptyToken(t *testing.T) {
	if _, err := HashRefreshToken(""); err == nil {
		t.Fatal("HashRefreshToken() error = nil, want error")
	}
}
