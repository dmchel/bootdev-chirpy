package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

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

func TestMakeJWT(t *testing.T) {
	userId := uuid.New()
	token, err := MakeJWT(userId, "test-secret", time.Second*30)
	if err != nil {
		t.Error("Unexpected error:", err)
		return
	}

	if len(token) == 0 {
		t.Error("Empty token returned")
		return
	}

	fmt.Println(token)

	actualId, err := ValidateJWT(token, "test-secret")
	if err != nil {
		t.Error("Unexpected error:", err)
		return
	}

	if actualId != userId {
		t.Error("User ID from token doesn't match initial value")
	}
}

func TestExpiredJWT(t *testing.T) {
	userId := uuid.New()
	token, err := MakeJWT(userId, "test-secret", time.Second*1)
	if err != nil {
		t.Error("Unexpected error:", err)
		return
	}

	if len(token) == 0 {
		t.Error("Empty token returned")
		return
	}

	fmt.Println(token)

	time.Sleep(2 * time.Second)

	_, err = ValidateJWT(token, "test-secret")
	if err == nil {
		t.Error("No error reported for an expired token")
	}
}
