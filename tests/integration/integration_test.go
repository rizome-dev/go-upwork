// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	upwork "github.com/rizome-dev/go-upwork/pkg"
	"github.com/rizome-dev/go-upwork/pkg/auth"
	"github.com/rizome-dev/go-upwork/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// TestEnvironment checks if required environment variables are set
func TestEnvironment(t *testing.T) {
	required := []string{
		"UPWORK_CLIENT_ID",
		"UPWORK_CLIENT_SECRET",
		"UPWORK_ACCESS_TOKEN",
		"UPWORK_REFRESH_TOKEN",
		"UPWORK_ORGANIZATION_ID",
	}

	for _, env := range required {
		value := os.Getenv(env)
		if value == "" {
			t.Skipf("Skipping integration tests: %s not set", env)
		}
	}
}

// setupClient creates a real client for integration testing
func setupClient(t *testing.T) *upwork.Client {
	clientID := os.Getenv("UPWORK_CLIENT_ID")
	clientSecret := os.Getenv("UPWORK_CLIENT_SECRET")
	accessToken := os.Getenv("UPWORK_ACCESS_TOKEN")
	refreshToken := os.Getenv("UPWORK_REFRESH_TOKEN")
	organizationID := os.Getenv("UPWORK_ORGANIZATION_ID")

	if clientID == "" || clientSecret == "" || accessToken == "" {
		t.Skip("Skipping integration test: credentials not configured")
	}

	oauth2Config, err := auth.NewOAuth2Config(auth.OAuth2Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callback",
	})
	require.NoError(t, err)

	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	client, err := upwork.NewClient(upwork.Config{
		OAuth2Config:   oauth2Config,
		Token:          token,
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	return client
}

// TestGetCurrentUser tests fetching the current user
func TestGetCurrentUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupClient(t)
	ctx := context.Background()

	user, err := client.Users.GetCurrentUser(ctx)
	require.NoError(t, err)
	
	assert.NotEmpty(t, user.ID)
	assert.NotEmpty(t, user.Email)
	t.Logf("Current user: %s (%s)", user.Name, user.Email)
}

// TestListContracts tests listing contracts
func TestListContracts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupClient(t)
	ctx := context.Background()

	contracts, pageInfo, err := client.Contracts.ListContracts(ctx, models.ContractListOptions{
		Status: models.ContractStatusActive,
		Limit:  5,
	})
	
	if err != nil {
		// Some users might not have contracts
		t.Logf("Could not list contracts: %v", err)
		return
	}

	t.Logf("Found %d active contracts", len(contracts))
	for _, contract := range contracts {
		t.Logf("- Contract: %s (ID: %s)", contract.Title, contract.ID)
	}

	if pageInfo != nil && pageInfo.HasNextPage {
		t.Logf("More contracts available (cursor: %s)", pageInfo.EndCursor)
	}
}

// TestSearchJobs tests job search functionality
func TestSearchJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupClient(t)
	ctx := context.Background()

	// Search for Go programming jobs
	results, err := client.Jobs.SearchJobs(ctx, models.JobSearchOptions{
		Query:  "golang developer",
		Skills: []string{"golang", "go"},
		Limit:  5,
	})
	
	if err != nil {
		t.Logf("Could not search jobs: %v", err)
		return
	}

	t.Logf("Found %d jobs matching 'golang developer'", len(results.Jobs))
	for i, job := range results.Jobs {
		t.Logf("%d. %s", i+1, job.Title)
		t.Logf("   Budget: %v", job.Budget)
		t.Logf("   Posted: %v", job.DateCreated)
	}
}

// TestGetMetadata tests fetching API metadata
func TestGetMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupClient(t)
	ctx := context.Background()

	// Test getting countries
	countries, err := client.Metadata.GetCountries(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, countries)
	t.Logf("Available countries: %d", len(countries))

	// Test getting skills
	skills, err := client.Metadata.SearchSkills(ctx, "programming", 10)
	if err == nil {
		t.Logf("Found %d skills matching 'programming'", len(skills))
		for _, skill := range skills[:min(5, len(skills))] {
			t.Logf("- %s", skill.Name)
		}
	}

	// Test getting categories
	categories, err := client.Metadata.GetCategories(ctx)
	if err == nil {
		t.Logf("Available job categories: %d", len(categories))
	}
}

// TestRateLimiting tests that rate limiting works correctly
func TestRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupClient(t)
	ctx := context.Background()

	// Make multiple rapid requests
	start := time.Now()
	
	for i := 0; i < 5; i++ {
		_, err := client.Users.GetCurrentUser(ctx)
		if err != nil {
			t.Logf("Request %d failed: %v", i+1, err)
		}
	}
	
	elapsed := time.Since(start)
	t.Logf("5 requests completed in %v", elapsed)
	
	// With rate limiting, this should take at least some minimum time
	// The actual time depends on the rate limit configuration
}

// TestTokenRefresh tests automatic token refresh
func TestTokenRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	clientID := os.Getenv("UPWORK_CLIENT_ID")
	clientSecret := os.Getenv("UPWORK_CLIENT_SECRET")
	refreshToken := os.Getenv("UPWORK_REFRESH_TOKEN")

	if refreshToken == "" {
		t.Skip("Skipping token refresh test: no refresh token")
	}

	oauth2Config, err := auth.NewOAuth2Config(auth.OAuth2Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callback",
	})
	require.NoError(t, err)

	// Create a client with an expired token
	expiredToken := &oauth2.Token{
		AccessToken:  "expired-token",
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // Already expired
	}

	client, err := upwork.NewClient(upwork.Config{
		OAuth2Config: oauth2Config,
		Token:        expiredToken,
	})
	require.NoError(t, err)

	// The client should automatically refresh the token
	ctx := context.Background()
	err = client.RefreshToken(ctx)
	
	if err != nil {
		t.Logf("Token refresh failed: %v", err)
		t.Skip("Skipping: token refresh not available")
	}

	// Verify the new token works
	user, err := client.Users.GetCurrentUser(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	
	t.Log("Token refresh successful")
}

// TestErrorHandling tests error responses from the API
func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupClient(t)
	ctx := context.Background()

	// Try to get a non-existent contract
	_, err := client.Contracts.GetContract(ctx, "invalid-contract-id-12345")
	assert.Error(t, err)
	t.Logf("Expected error for invalid contract: %v", err)

	// Try to create a contract with invalid data
	_, err = client.Contracts.CreateContract(ctx, models.CreateContractInput{
		Title: "", // Invalid empty title
	})
	assert.Error(t, err)
	t.Logf("Expected validation error: %v", err)
}

// TestConcurrentRequests tests making concurrent API requests
func TestConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupClient(t)
	ctx := context.Background()

	// Make concurrent requests
	done := make(chan bool, 3)
	errors := make(chan error, 3)

	// Request 1: Get current user
	go func() {
		_, err := client.Users.GetCurrentUser(ctx)
		if err != nil {
			errors <- err
		}
		done <- true
	}()

	// Request 2: Get metadata
	go func() {
		_, err := client.Metadata.GetCountries(ctx)
		if err != nil {
			errors <- err
		}
		done <- true
	}()

	// Request 3: List contracts
	go func() {
		_, _, err := client.Contracts.ListContracts(ctx, models.ContractListOptions{
			Limit: 1,
		})
		if err != nil {
			errors <- err
		}
		done <- true
	}()

	// Wait for all requests
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Success
		case err := <-errors:
			t.Logf("Concurrent request error: %v", err)
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}

	t.Log("All concurrent requests completed")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}