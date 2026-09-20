package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestMintVerifyAccessToken_RoundTrip(t *testing.T) {
	tok, err := MintAccessToken("test-secret", "user-1", 15*time.Minute)
	require.NoError(t, err)
	sub, err := VerifyAccessToken("test-secret", tok)
	require.NoError(t, err)
	assert.Equal(t, "user-1", sub)
}

func TestVerifyAccessToken_Rejects(t *testing.T) {
	tok, err := MintAccessToken("test-secret", "user-1", 15*time.Minute)
	require.NoError(t, err)

	_, err = VerifyAccessToken("wrong-secret", tok)
	require.Error(t, err)

	expired, err := MintAccessToken("test-secret", "user-1", -time.Minute)
	require.NoError(t, err)
	_, err = VerifyAccessToken("test-secret", expired)
	require.Error(t, err)

	// Unsigned (alg=none) token must be rejected.
	none := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "user-1"})
	raw, err := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	_, err = VerifyAccessToken("test-secret", raw)
	require.Error(t, err)

	_, err = VerifyAccessToken("test-secret", "not-a-token")
	require.Error(t, err)
}

func TestNewRefreshToken_EntropyAndHash(t *testing.T) {
	raw1, hash1, err := NewRefreshToken()
	require.NoError(t, err)
	raw2, hash2, err := NewRefreshToken()
	require.NoError(t, err)

	assert.NotEqual(t, raw1, raw2)
	assert.NotEqual(t, hash1, hash2)

	b, err := base64.RawURLEncoding.DecodeString(raw1)
	require.NoError(t, err)
	assert.Len(t, b, 32)
	assert.Equal(t, HashRefreshToken(raw1), hash1)
	assert.True(t, RefreshTokenMatches(hash1, raw1))
	assert.False(t, RefreshTokenMatches(hash1, raw2))
	assert.False(t, RefreshTokenMatches("short", raw1))
}

func TestDummyPasswordHash_IsValidCost12(t *testing.T) {
	cost, err := bcrypt.Cost([]byte(dummyPasswordHash))
	require.NoError(t, err)
	assert.Equal(t, 12, cost)
	assert.True(t, strings.HasPrefix(dummyPasswordHash, "$2a$12$"))
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte("anything-at-all")))
}
