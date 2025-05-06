package util

import (
	"errors"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// setup sets the environment variable for the secret key before tests run.
func setupJWTTest(t *testing.T) func(t *testing.T) {
	// Set a temporary secret key for testing
	originalSecret := os.Getenv("SECRET_KEY")
	os.Setenv("SECRET_KEY", "test-secret-key-for-util") // Use a specific key for this test

	// Return a teardown function to restore the original value
	return func(t *testing.T) {
		os.Setenv("SECRET_KEY", originalSecret)
	}
}

func TestGenerateToken(t *testing.T) {
	// Setup
	teardown := setupJWTTest(t)
	defer teardown(t) // Ensure teardown runs even if test panics

	jwtService := NewJwtImpl()
	userID := "123"
	role := "teacher"

	// Test token generation
	token, err := jwtService.GenerateToken(userID, role)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Optional: Decode and check claims (without validation here)
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	assert.NoError(t, err, "Should be able to parse the generated token")
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok, "Claims should be MapClaims")
	assert.Equal(t, userID, claims["userID"])
	assert.Equal(t, role, claims["role"])
	// Check expiry is roughly correct (within a reasonable window)
	expFloat, ok := claims["exp"].(float64)
	assert.True(t, ok, "Expiry should be a number")
	expectedExp := time.Now().Add(TOKEN_EXPIRED_TIME).Unix()
	assert.InDelta(t, expectedExp, int64(expFloat), 5, "Expiry time should be close to expected") // Allow 5s delta
}

func TestValidateToken(t *testing.T) {
	// Setup
	teardown := setupJWTTest(t)
	defer teardown(t)

	jwtService := NewJwtImpl()
	userID := "123"
	role := "teacher"

	// Generate a token first
	token, err := jwtService.GenerateToken(userID, role)
	require.NoError(t, err) // Use require as validation depends on valid generation

	// Test validation
	claims, err := jwtService.ValidateToken(token)

	// Assertions
	assert.NoError(t, err)
	require.NotNil(t, claims) // Use require as subsequent checks depend on non-nil claims
	assert.Equal(t, userID, claims["userID"])
	assert.Equal(t, role, claims["role"])
}

func TestValidateExpiredToken(t *testing.T) {
	// Setup
	teardown := setupJWTTest(t)
	defer teardown(t)

	jwtService := NewJwtImpl()
	secretKey := []byte(os.Getenv("SECRET_KEY")) // Get the key set in setup

	// Create an expired token manually
	expiredTime := time.Now().Add(-time.Hour) // 1 hour in the past
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": "123",
		"role":   "student",
		"exp":    expiredTime.Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	require.NoError(t, err)

	// Validate the expired token
	claims, err := jwtService.ValidateToken(tokenString)

	// Should get an error for expired token
	assert.Error(t, err)
	assert.Nil(t, claims)
	// Check for specific JWT error if possible (optional but good)
	assert.True(t, errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenInvalidClaims), "Expected ErrTokenExpired or ErrTokenInvalidClaims")

}

func TestValidateInvalidTokenSignature(t *testing.T) {
	// Setup
	teardown := setupJWTTest(t)
	defer teardown(t)

	jwtService := NewJwtImpl()
	secretKey := []byte(os.Getenv("SECRET_KEY"))
	wrongSecretKey := []byte("different-secret-key")

	// Generate token with correct key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": "123",
		"role":   "student",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	_, err := token.SignedString(secretKey)
	require.NoError(t, err)

	// Try to validate with the wrong key (simulated by tampering or wrong key)
	// We can't directly validate with a wrong key in ValidateToken,
	// but we can test with a token signed by a different key.
	tokenSignedWithWrongKey, err := token.SignedString(wrongSecretKey)
	require.NoError(t, err)

	claims, err := jwtService.ValidateToken(tokenSignedWithWrongKey)

	// Should get an error for invalid signature
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.True(t, errors.Is(err, jwt.ErrSignatureInvalid), "Expected ErrSignatureInvalid")
}

func TestValidateMalformedToken(t *testing.T) {
	// Setup
	teardown := setupJWTTest(t)
	defer teardown(t)

	jwtService := NewJwtImpl()

	// Test with invalid token string
	malformedToken := "this.is.not.a.valid.jwt"
	claims, err := jwtService.ValidateToken(malformedToken)

	// Should get an error
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_MissingClaims(t *testing.T) {
	// Setup
	teardown := setupJWTTest(t)
	defer teardown(t)

	jwtService := NewJwtImpl()
	secretKey := []byte(os.Getenv("SECRET_KEY"))

	// Create a token *without* userID or role
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		// Missing "userID" and "role"
	})
	tokenString, err := token.SignedString(secretKey)
	require.NoError(t, err)

	// Validate the token
	claims, err := jwtService.ValidateToken(tokenString)

	// Validation itself might pass if expiry is okay,
	// but the claims map will be missing expected fields.
	// The current ValidateToken doesn't explicitly check for missing required claims,
	// it just returns the parsed claims. Middleware using this should check.
	assert.NoError(t, err) // Expect no parsing/validation error based on expiry/signature
	require.NotNil(t, claims)
	_, userIDExists := claims["userID"]
	_, roleExists := claims["role"]
	assert.False(t, userIDExists, "userID claim should be missing")
	assert.False(t, roleExists, "role claim should be missing")
}
