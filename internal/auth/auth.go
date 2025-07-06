package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const (
	TokenTypeAccess TokenType = "chirpy-access"
)

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func HashPassword(password string) (string, error) {
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error: failed to hash password: %s", err)
	}

	return string(hashed_password), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("error: passwords don't match: %s", err)
	}

	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// creates new jwt token with our constant issuer
	// and the user's ID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	// signs and returns the token based on our secret
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}

	// parse and validate the token
	// anonymous function provides the secret for signature verification
	// automatically validates expiration time and signature
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (any, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}

	// extract userID
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	// extracts and checks the issuer is us
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	// convert userID string to uuid and return
	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")

	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("malformed authorization header")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func MakeRefreshToken() (string, error) {
	randomData := make([]byte, 32)
	_, err := rand.Read(randomData) // makes random 32-byte data
	if err != nil {
		return "", errors.New("error: failed to make refresh token")
	}

	refreshToken := hex.EncodeToString(randomData)
	return refreshToken, nil
}
