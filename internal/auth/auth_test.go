package auth

import (
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	type testCase struct {
		name    string
		headers http.Header
		want    string
		wantErr bool
	}

	tests := []testCase{
		{
			name:    "Valid bearer token",
			headers: http.Header{"Authorization": {"Bearer my_valid_token"}},
			want:    "my_valid_token",
			wantErr: false,
		},
		{
			name:    "No authorization header",
			headers: http.Header{"Content-Type": {"application/json"}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty headers",
			headers: http.Header{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Authorization header without Bearer prefix",
			headers: http.Header{"Authorization": {"just_a_token"}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Authorization header with different prefix (Basic)",
			headers: http.Header{"Authorization": {"Basic YWRtaW46cGFzc3dvcmQ="}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Authorization header with empty token",
			headers: http.Header{"Authorization": {"Bearer "}},
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBearerToken(tt.headers)

			// Assert error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetBearerToken() got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetBearerToken2(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name: "Valid Bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer valid_token"},
			},
			wantToken: "valid_token",
			wantErr:   false,
		},
		{
			name:      "Missing Authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "Malformed Authorization header",
			headers: http.Header{
				"Authorization": []string{"InvalidBearer token"},
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}
