package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/petieclark/pews/internal/auth"
	"github.com/petieclark/pews/internal/billing"
	"github.com/petieclark/pews/internal/checkins"
	"github.com/petieclark/pews/internal/communication"
	"github.com/petieclark/pews/internal/config"
	"github.com/petieclark/pews/internal/giving"
	"github.com/petieclark/pews/internal/groups"
	"github.com/petieclark/pews/internal/module"
	"github.com/petieclark/pews/internal/people"
	"github.com/petieclark/pews/internal/router"
	"github.com/petieclark/pews/internal/services"
	"github.com/petieclark/pews/internal/sms"
	"github.com/petieclark/pews/internal/streaming"
	"github.com/petieclark/pews/internal/tenant"
)

// testServer wraps httptest.Server with helper methods
type testServer struct {
	*httptest.Server
	authToken string
}

// newTestServer creates a new test server with all routes configured
func newTestServer() *testServer {
	pool := getTestPool()

	// Initialize services
	authService := auth.NewService(pool, jwtSecret)
	tenantService := tenant.NewService(pool)
	moduleService := module.NewService(pool)
	billingService := billing.NewService(pool, "", "", frontendURL)
	peopleService := people.NewService(pool)
	groupsService := groups.NewService(pool)
	servicesService := services.NewService(pool)
	givingService := giving.NewService(pool)
	givingStripeService := giving.NewStripeService(pool, "", frontendURL)
	streamingService := streaming.NewService(pool)
	communicationService := communication.NewService(pool)
	checkinsService := checkins.NewService(pool)
	smsService := sms.NewService(pool)

	// Initialize handlers
	authHandler := auth.NewHandler(authService, tenantService, billingService)
	tenantHandler := tenant.NewHandler(tenantService)
	moduleHandler := module.NewHandler(moduleService)
	billingHandler := billing.NewHandler(billingService)
	peopleHandler := people.NewHandler(peopleService)
	groupsHandler := groups.NewHandler(groupsService)
	servicesHandler := services.NewHandler(servicesService)
	givingHandler := giving.NewHandler(givingService, givingStripeService)
	streamingHandler := streaming.NewHandler(streamingService)
	communicationHandler := communication.NewHandler(communicationService)
	checkinsHandler := checkins.NewHandler(checkinsService)
	smsHandler := sms.NewHandler(smsService)

	// Setup router
	r := router.New(
		authHandler,
		authService,
		tenantHandler,
		moduleHandler,
		billingHandler,
		peopleHandler,
		groupsHandler,
		servicesHandler,
		givingHandler,
		streamingHandler,
		communicationHandler,
		checkinsHandler,
		smsHandler,
		"test-webhook-secret",
		"test-giving-webhook-secret",
		frontendURL,
	)

	return &testServer{
		Server: httptest.NewServer(r),
	}
}

// doRequest performs an HTTP request and returns the response
func (ts *testServer) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, ts.URL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if ts.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+ts.authToken)
	}

	return http.DefaultClient.Do(req)
}

// get performs a GET request
func (ts *testServer) get(path string) (*http.Response, error) {
	return ts.doRequest("GET", path, nil)
}

// post performs a POST request
func (ts *testServer) post(path string, body interface{}) (*http.Response, error) {
	return ts.doRequest("POST", path, body)
}

// put performs a PUT request
func (ts *testServer) put(path string, body interface{}) (*http.Response, error) {
	return ts.doRequest("PUT", path, body)
}

// delete performs a DELETE request
func (ts *testServer) delete(path string) (*http.Response, error) {
	return ts.doRequest("DELETE", path, nil)
}

// setAuthToken sets the authentication token for subsequent requests
func (ts *testServer) setAuthToken(token string) {
	ts.authToken = token
}

// register creates a new tenant and admin user, returns the auth token
func (ts *testServer) register(tenantName, email, password string) (string, error) {
	resp, err := ts.post("/api/auth/register", map[string]string{
		"tenant_name": tenantName,
		"email":       email,
		"password":    password,
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("register failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["token"].(string)
	if !ok {
		return "", fmt.Errorf("no token in response")
	}

	return token, nil
}

// login authenticates a user and returns the auth token
func (ts *testServer) login(tenantSlug, email, password string) (string, error) {
	resp, err := ts.post("/api/auth/login", map[string]string{
		"tenant_slug": tenantSlug,
		"email":       email,
		"password":    password,
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["token"].(string)
	if !ok {
		return "", fmt.Errorf("no token in response")
	}

	return token, nil
}

// decodeJSON decodes a JSON response body into the given interface
func decodeJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

// readBody reads the response body as a string
func readBody(resp *http.Response) string {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

// assertStatusCode fails the test if the response status doesn't match expected
func assertStatusCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body := readBody(resp)
		t.Fatalf("Expected status %d, got %d. Body: %s", expected, resp.StatusCode, body)
	}
}

// assertStatusCodeOKOrCreated checks for either 200 or 201 (useful for create endpoints)
func assertStatusCodeOKOrCreated(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body := readBody(resp)
		t.Fatalf("Expected status 200 or 201, got %d. Body: %s", resp.StatusCode, body)
	}
}

// assertNoError fails the test if err is not nil
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// assertEqual fails the test if expected != actual
func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// assertNotEqual fails the test if expected == actual
func assertNotEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected == actual {
		t.Fatalf("Expected values to be different, but both were %v", expected)
	}
}

// cleanupBeforeTest truncates all tables before each test
func cleanupBeforeTest(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	cleanupDatabase(ctx)
}

// loadConfig returns a test configuration
func loadConfig() *config.Config {
	return &config.Config{
		DatabaseURL:         testDBURL,
		JWTSecret:           jwtSecret,
		StripeSecretKey:     "",
		StripeWebhookSecret: "test-webhook-secret",
		StripePriceID:       "",
		Port:                "8080",
		FrontendURL:         frontendURL,
	}
}
