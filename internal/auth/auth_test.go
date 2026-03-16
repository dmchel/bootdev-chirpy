package auth

import "testing"

func TestHashPassword(t *testing.T) {
	password := "test123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Error("Unexpected error", err)
		return
	}
	if len(hash) == 0 {
		t.Error("Empty hash returned")
		return
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Error("Unexpected error", err)
		return
	}
	if !match {
		t.Error("Password Hash mismatch")
	}
}
