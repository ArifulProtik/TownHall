package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const BcryptCost = 12

const RefreshTokenBytes = 32

// Compared on unknown-email login so failures take as long as wrong passwords.
const dummyPasswordHash = "$2a$12$RMU9l7T/F9JUd0HsvLdQGOuKF.Zp5PUDS1H89./B9ibJ0ip9yZyFa"

// MintAccessToken issues an HS256 access token for userID valid for ttl.
func MintAccessToken(secret, userID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": now.Add(ttl).Unix(),
		"iat": now.Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func VerifyAccessToken(secret, token string) (string, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return "", errors.New("invalid token claims")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", errors.New("missing subject")
	}
	return sub, nil
}

// NewRefreshToken returns a raw token for the cookie plus its hash for storage.
func NewRefreshToken() (raw, hash string, err error) {
	var b [RefreshTokenBytes]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b[:])
	return raw, HashRefreshToken(raw), nil
}

func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// Production lookup is by DB equality on the hash; this is for re-checks and tests.
func RefreshTokenMatches(storedHash, rawCandidate string) bool {
	candidate := HashRefreshToken(rawCandidate)
	if len(storedHash) != len(candidate) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(candidate)) == 1
}
