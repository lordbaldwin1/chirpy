package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "mySecurePassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "p@ssw0rd!#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if hash == "" {
					t.Error("HashPassword() returned empty hash")
				}

				if hash == tt.password {
					t.Error("HashPassword() returned unhashed password")
				}

				err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
				if err != nil {
					t.Errorf("Generated hash is invalid: %v", err)
				}
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	testPassword := "testPassword123"
	validHash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate test hash: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "valid password and hash",
			password: testPassword,
			hash:     string(validHash),
			wantErr:  false,
		},
		{
			name:     "invalid password",
			password: "wrongPassword",
			hash:     string(validHash),
			wantErr:  true,
		},
		{
			name:     "empty password with valid hash",
			password: "",
			hash:     string(validHash),
			wantErr:  true,
		},
		{
			name:     "valid password with empty hash",
			password: testPassword,
			hash:     "",
			wantErr:  true,
		},
		{
			name:     "invalid hash format",
			password: testPassword,
			hash:     "invalid-hash-format",
			wantErr:  true,
		},
		{
			name:     "both empty",
			password: "",
			hash:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	passwords := []string{
		"simplePassword",
		"complex!P@ssw0rd#123",
		"",
	}

	for _, password := range passwords {
		t.Run("password_"+password, func(t *testing.T) {
			hash, err := HashPassword(password)
			if err != nil {
				t.Fatalf("HashPassword() failed: %v", err)
			}

			err = CheckPasswordHash(password, hash)
			if err != nil {
				t.Errorf("CheckPasswordHash() failed for correct password: %v", err)
			}

			wrongPassword := password + "wrong"
			err = CheckPasswordHash(wrongPassword, hash)
			if err == nil {
				t.Error("CheckPasswordHash() should have failed for wrong password")
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: "secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}
