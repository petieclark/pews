package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthRegister(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	t.Run("successful registration", func(t *testing.T) {
		resp, err := ts.post("/api/auth/register", map[string]string{
			"tenant_name": "Test Church",
			"email":       "admin@testchurch.com",
			"password":    "password123",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusOK)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		// Verify response contains required fields
		if result["token"] == nil {
			t.Fatal("Expected token in response")
		}
		if result["tenant_id"] == nil {
			t.Fatal("Expected tenant_id in response")
		}
		if result["email"] != "admin@testchurch.com" {
			t.Fatalf("Expected email admin@testchurch.com, got %v", result["email"])
		}
		if result["role"] != "admin" {
			t.Fatalf("Expected role admin, got %v", result["role"])
		}

		// Verify tenant was created
		ctx := context.Background()
		var tenantName string
		err = testDB.Pool.QueryRow(ctx, "SELECT name FROM tenants WHERE id = $1", result["tenant_id"]).Scan(&tenantName)
		assertNoError(t, err)
		assertEqual(t, "Test Church", tenantName)
	})

	t.Run("missing required fields", func(t *testing.T) {
		resp, err := ts.post("/api/auth/register", map[string]string{
			"tenant_name": "Test Church",
			"email":       "",
			"password":    "password123",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusBadRequest)
	})
}

func TestAuthLogin(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	// Register a test user first
	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	if token == "" {
		t.Fatal("Expected token from registration")
	}

	t.Run("successful login", func(t *testing.T) {
		resp, err := ts.post("/api/auth/login", map[string]string{
			"tenant_slug": "test-church",
			"email":       "admin@testchurch.com",
			"password":    "password123",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusOK)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if result["token"] == nil {
			t.Fatal("Expected token in response")
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		resp, err := ts.post("/api/auth/login", map[string]string{
			"tenant_slug": "test-church",
			"email":       "admin@testchurch.com",
			"password":    "wrongpassword",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusUnauthorized)
	})

	t.Run("invalid tenant", func(t *testing.T) {
		resp, err := ts.post("/api/auth/login", map[string]string{
			"tenant_slug": "nonexistent",
			"email":       "admin@testchurch.com",
			"password":    "password123",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusUnauthorized)
	})

	t.Run("missing fields", func(t *testing.T) {
		resp, err := ts.post("/api/auth/login", map[string]string{
			"tenant_slug": "test-church",
			"email":       "",
			"password":    "password123",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusBadRequest)
	})
}

func TestAuthProtectedRoutes(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	t.Run("access without token returns 401", func(t *testing.T) {
		resp, err := ts.get("/api/people")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusUnauthorized)
	})

	t.Run("access with valid token succeeds", func(t *testing.T) {
		token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
		assertNoError(t, err)
		
		ts.setAuthToken(token)
		resp, err := ts.get("/api/people")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusOK)
	})

	t.Run("access with invalid token returns 401", func(t *testing.T) {
		ts.setAuthToken("invalid.token.here")
		resp, err := ts.get("/api/people")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusUnauthorized)
	})
}

func TestAuthExpiredToken(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	// Create an expired token manually
	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)

	// Parse the token to get claims
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	assertNoError(t, err)

	claims := parsedToken.Claims.(jwt.MapClaims)
	
	// Create a new token with past expiration
	expiredClaims := jwt.MapClaims{
		"user_id":   claims["user_id"],
		"tenant_id": claims["tenant_id"],
		"email":     claims["email"],
		"role":      claims["role"],
		"exp":       time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
	}

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	tokenString, err := expiredToken.SignedString([]byte(jwtSecret))
	assertNoError(t, err)

	// Try to access with expired token
	ts.setAuthToken(tokenString)
	resp, err := ts.get("/api/people")
	assertNoError(t, err)
	assertStatusCode(t, resp, http.StatusUnauthorized)
}

func TestAuthLogout(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)

	ts.setAuthToken(token)

	t.Run("logout succeeds", func(t *testing.T) {
		resp, err := ts.post("/api/auth/logout", nil)
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusOK)
	})
}
