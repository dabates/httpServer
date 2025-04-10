package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/dabates/httpServer/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func MakeJWT(userid uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userid.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(token string, tokenSecret string) (uuid.UUID, error) {
	t, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	claims := t.Claims.(*jwt.RegisteredClaims)
	return uuid.Parse(claims.Subject)
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("authorization header is missing or empty")
	}

	if strings.Contains(strings.ToLower(auth), "bearer") == false {
		return "", fmt.Errorf("authorization header is missing the 'Bearer' prefix")
	}

	auth = strings.TrimSpace(strings.Replace(auth, "Bearer ", "", 1))

	return auth, nil
}

func MakeRefreshToken(userId uuid.UUID, db *database.Queries) (string, error) {
	key := make([]byte, 32)
	rand.Read(key)

	hexKey := fmt.Sprintf("%x", key)
	now := time.Now()

	_, err := db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		ExpiresAt: sql.NullTime{
			Time:  now.AddDate(0, 0, 60),
			Valid: true,
		},
		Token:  hexKey,
		UserID: userId,
	})

	if err != nil {
		return "", err
	}

	return hexKey, nil
}

func ValidateRefreshToken(token string, db *database.Queries) (bool, error) {
	tokenRec, err := db.GetUserFromRefreshToken(context.Background(), token)
	if err != nil {
		return false, err
	}

	if tokenRec.RevokedAt.Valid || tokenRec.ExpiresAt.Time.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

func GetApiKey(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("authorization header is missing or empty")
	}

	if strings.Contains(strings.ToLower(auth), "apikey") == false {
		return "", fmt.Errorf("authorization header is missing the 'ApiKey' prefix")
	}

	auth = strings.TrimSpace(strings.Replace(auth, "ApiKey ", "", 1))

	return auth, nil
}
