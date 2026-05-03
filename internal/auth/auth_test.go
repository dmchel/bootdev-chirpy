package auth

import (
	"fmt"
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	headers := make(http.Header)
	headers.Add("Authorization", "Bearer 12345678")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Error("Unexpected error:", err)
		return
	}

	if token != "12345678" {
		t.Errorf("Unexpected token returned")
	}
}

func TestGetApiKey(t *testing.T) {
	headers := make(http.Header)
	headers.Add("Authorization", "ApiKey adcdefgh12345")

	apiKey, err := GetApiKey(headers)
	if err != nil {
		t.Error("Unexpected error:", err)
		return
	}

	if apiKey != "adcdefgh12345" {
		t.Errorf("Unexpected api key returned: got %s, expect %s", apiKey, "adcdefgh12345")
	}
}
