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

// BcryptCost is the bcrypt cost applied to all password hashing.
const BcryptCost = 12

// RefreshTokenBytes is the crypto/rand entropy size of raw refresh tokens.
const RefreshTokenBytes = 32

// dummyPasswordHash is a valid cost-12 bcrypt hash compared against on
// unknown-email login so timing matches the wrong-password path.
const dummyPasswordHash = "$2a$12$RMU9l7T/F9JUd0HsvLdQGOuKF.Zp5PUDS1H89./B9ibJ0ip9yZyFa" //nolint:gosec // dummy bcrypt hash for timing-safe login, not a credential

// MintAccessToken issues an HS256 access JWT for userID valid for ttl.
func MintAccessToken(secret, userID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": now.Add(ttl).Unix(),
		"iat": now.Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// VerifyAccessToken validates token against secret and returns the subject.
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

// NewRefreshToken generates a raw opaque token plus its SHA-256 storage hash.
// Only the hash may be persisted or logged; the raw value goes to the cookie.
func NewRefreshToken() (raw, hash string, err error) {
	var b [RefreshTokenBytes]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b[:])
	return raw, HashRefreshToken(raw), nil
}

// HashRefreshToken returns the hex SHA-256 of a raw refresh token.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// RefreshTokenMatches constant-time compares a stored hash to a raw candidate.
// Production lookup is by DB equality on the SHA-256 hash (256-bit random
// tokens have no practical timing leak); this helper exists for explicit
// re-comparison and unit tests.
func RefreshTokenMatches(storedHash, rawCandidate string) bool {
	candidate := HashRefreshToken(rawCandidate)
	if len(storedHash) != len(candidate) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(candidate)) == 1
}
