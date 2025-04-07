package auth

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	password := "securepassword123"

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the hashed password is valid
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		t.Fatalf("Password hash validation failed: %v", err)
	}

	t.Log("Password hashing and validation succeeded")
}

func TestCheckPasswordHash(t *testing.T) {
	password := "securepassword123"

	// Create a hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Unable to generate password hash: %v", err)
	}

	// Test the CheckPasswordHash function
	check := CheckPasswordHash(password, string(hashedPassword))
	if !check {
		t.Fatalf("Expected correct password to match the hash, but it didn't")
	}

	// Test with the wrong password
	checkWrong := CheckPasswordHash("wrongpassword", string(hashedPassword))
	if checkWrong {
		t.Fatalf("Expected wrong password not to match the hash, but it did")
	}

	t.Log("CheckPasswordHash passed for both correct and incorrect passwords")
}

func TestMakeJWTAndValidateJWT(t *testing.T) {
	tokenSecret := "supersecretkey"
	userID := uuid.New()
	expiresIn := time.Minute * 5

	// Generate a JWT
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Validate the generated JWT
	parsedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	// Check if the userID matches
	if parsedUserID != userID {
		t.Fatalf("Expected UserID %v, got %v", userID, parsedUserID)
	}

	// Validate an expired token (simulate by creating a token that expired)
	expiredToken, err := MakeJWT(userID, tokenSecret, -time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate expired JWT: %v", err)
	}
	_, err = ValidateJWT(expiredToken, tokenSecret)
	if err == nil {
		t.Fatal("Expected error when validating an expired token, but got none")
	}

	t.Log("MakeJWT and ValidateJWT passed for valid and expired tokens")
}

func TestGetBearerToken(t *testing.T) {
	// Case 1: Valid Bearer token
	headers := http.Header{}
	headers.Set("Authorization", "Bearer TOKEN_STRING")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if token != "TOKEN_STRING" {
		t.Fatalf("Expected token 'TOKEN_STRING', got '%s'", token)
	}

	// Case 2: Authorization header with extra whitespace
	headers.Set("Authorization", "  Bearer    TOKEN_STRING   ") // Intentionally messy
	token, err = GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if token != "TOKEN_STRING" {
		t.Fatalf("Expected token 'TOKEN_STRING', got '%s'", token)
	}

	// Case 3: Authorization header without "Bearer" prefix
	headers.Set("Authorization", "TOKEN_STRING")
	_, err = GetBearerToken(headers)
	if err == nil {
		t.Fatal("Expected an error when Authorization header is missing the 'Bearer' prefix")
	}

	// Case 4: Missing Authorization header
	headers = http.Header{}
	token, err = GetBearerToken(headers)
	if err == nil {
		t.Fatal("Expected an error when Authorization header is missing, but got no error")
	}

	// Case 5: Empty Authorization header
	headers.Set("Authorization", "")
	token, err = GetBearerToken(headers)
	if err == nil {
		t.Fatal("Expected an error when Authorization header is empty, but got no error")
	}

	t.Log("GetBearerToken passed all test cases")
}
