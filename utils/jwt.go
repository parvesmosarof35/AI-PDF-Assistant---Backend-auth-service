package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID string, email string) (string, error) {
	secret := os.Getenv("JWT_ACCESS_SECRET")
	if secret == "" {
		secret = "fallback_secret_key"
	}

	// 240h = 10 days
	expirationTime := time.Now().Add(240 * time.Hour)

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   expirationTime.Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
