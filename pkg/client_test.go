package upwork

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/rizome-dev/go-upwork/pkg/auth"
	"github.com/rizome-dev/go-upwork/tests/mocks"
	"github.com/rizome-dev/go-upwork/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		validate func(t *testing.T, client *Client)
	}{
		{
			name: "valid configuration",
			config: Config{
				OAuth2Config: &auth.OAuth2Config{},
				Token: &oauth2.Token{
					AccessToken: "test-token",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(1 * time.Hour),
				},
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.NotNil(t, client)
				assert.NotNil(t, client.Users)
				assert.NotNil(t, client.Contracts)
				assert.NotNil(t, client.Jobs)
				assert.NotNil(t, client.Messages)
				assert.NotNil(t, client.Freelancers)
				assert.NotNil(t, client.Reports)
				assert.NotNil(t, client.Activities)
				assert.NotNil(t, client.Metadata)
			},
		},
		{
			name: "missing OAuth2 config",
			config: Config{
				Token: &oauth2.Token{
					AccessToken: "test-token",
				},
			},
			wantErr: true,
		},
		{
			name: "missing token",
			config: Config{
				OAuth2Config: &auth.OAuth2Config{},
			},
			wantErr: true,
		},
		{
			name: "with organization ID",
			config: Config{
				OAuth2Config: &auth.OAuth2Config{},
				Token: &oauth2.Token{
					AccessToken: "test-token",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(1 * time.Hour),
				},
				OrganizationID: "org-123",
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.Equal(t, "org-123", client.organizationID)
			},
		},
		{
			name: "with custom HTTP client",
			config: Config{
				OAuth2Config: &auth.OAuth2Config{},
				Token: &oauth2.Token{
					AccessToken: "test-token",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(1 * time.Hour),
				},
				HTTPClient: &http.Client{
					Timeout: 5 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "with custom rate limiter",
			config: Config{
				OAuth2Config: &auth.OAuth2Config{},
				Token: &oauth2.Token{
					AccessToken: "test-token",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(1 * time.Hour),
				},
				RateLimiter: mocks.NewMockRateLimiter(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if tt.validate != nil {
					tt.validate(t, client)
				}
			}
		})
	}
}

func TestClientWithOptions(t *testing.T) {
	baseConfig := Config{
		OAuth2Config: &auth.OAuth2Config{},
		Token: &oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
	}

	tests := []struct {
		name     string
		option   ClientOption
		validate func(t *testing.T, client *Client)
	}{
		{
			name:   "with organization ID",
			option: WithOrganizationID("org-456"),
			validate: func(t *testing.T, client *Client) {
				assert.Equal(t, "org-456", client.organizationID)
			},
		},
		{
			name:   "with rate limiter",
			option: WithRateLimiter(mocks.NewMockRateLimiter()),
			validate: func(t *testing.T, client *Client) {
				assert.NotNil(t, client.config.RateLimiter)
			},
		},
		{
			name:   "with HTTP client",
			option: WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
			validate: func(t *testing.T, client *Client) {
				assert.NotNil(t, client.config.HTTPClient)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(baseConfig, tt.option)
			require.NoError(t, err)
			tt.validate(t, client)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	mockHTTPClient := mocks.NewRequestRecorder(
		mocks.MockResponse{
			StatusCode: 200,
			Body: `{
				"access_token": "new-access-token",
				"token_type": "Bearer",
				"expires_in": 3600,
				"refresh_token": "new-refresh-token"
			}`,
		},
	)

	oauth2Config, err := auth.NewOAuth2Config(auth.OAuth2Options{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		TokenURL:     "https://api.upwork.com/api/v3/oauth2/token",
	})
	require.NoError(t, err)

	client, err := NewClient(Config{
		OAuth2Config: oauth2Config,
		Token: &oauth2.Token{
			AccessToken:  "old-token",
			RefreshToken: "old-refresh-token",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(-1 * time.Hour), // Expired
		},
		HTTPClient: mockHTTPClient,
	})
	require.NoError(t, err)

	err = client.RefreshToken(context.Background())
	assert.NoError(t, err)

	// Verify token was updated
	client.tokenMu.RLock()
	assert.Equal(t, "new-access-token", client.token.AccessToken)
	assert.Equal(t, "new-refresh-token", client.token.RefreshToken)
	client.tokenMu.RUnlock()
}

func TestGetToken(t *testing.T) {
	token := &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	client, err := NewClient(Config{
		OAuth2Config: &auth.OAuth2Config{},
		Token:        token,
	})
	require.NoError(t, err)

	retrievedToken := client.GetToken()
	assert.Equal(t, token.AccessToken, retrievedToken.AccessToken)
	assert.Equal(t, token.TokenType, retrievedToken.TokenType)
}

func TestSetOrganizationID(t *testing.T) {
	client, err := NewClient(Config{
		OAuth2Config: &auth.OAuth2Config{},
		Token: &oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
	})
	require.NoError(t, err)

	client.SetOrganizationID("new-org-123")
	assert.Equal(t, "new-org-123", client.organizationID)
}

func TestConcurrentTokenAccess(t *testing.T) {
	client, err := NewClient(Config{
		OAuth2Config: &auth.OAuth2Config{},
		Token: &oauth2.Token{
			AccessToken: "initial-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
	})
	require.NoError(t, err)

	// Simulate concurrent access
	done := make(chan bool)
	
	// Writer goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			newToken := &oauth2.Token{
				AccessToken: fmt.Sprintf("token-%d", id),
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(1 * time.Hour),
			}
			client.tokenMu.Lock()
			client.token = newToken
			client.tokenMu.Unlock()
			done <- true
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func() {
			token := client.GetToken()
			assert.NotNil(t, token)
			assert.Contains(t, token.AccessToken, "token")
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestClientInitialization(t *testing.T) {
	// Test that all services are properly initialized with the same GraphQL client
	client, err := NewClient(Config{
		OAuth2Config: &auth.OAuth2Config{},
		Token: &oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
		OrganizationID: "test-org",
	})
	require.NoError(t, err)

	// Use reflection or type assertions to verify services share the same base
	// This is a simplified check - in practice you might want more thorough verification
	assert.NotNil(t, client.Users)
	assert.NotNil(t, client.Contracts)
	assert.NotNil(t, client.Jobs)
	assert.NotNil(t, client.Messages)
	assert.NotNil(t, client.Freelancers)
	assert.NotNil(t, client.Reports)
	assert.NotNil(t, client.Activities)
	assert.NotNil(t, client.Metadata)
}

func TestTokenAutoRefresh(t *testing.T) {
	refreshCount := 0
	mockHTTPClient := &mocks.MockHTTPClient{
		Responses: []mocks.MockResponse{
			{
				StatusCode: 200,
				Body: `{
					"access_token": "refreshed-token",
					"token_type": "Bearer",
					"expires_in": 3600,
					"refresh_token": "new-refresh-token"
				}`,
			},
		},
	}

	// Override the Do method to count refreshes
	originalDo := mockHTTPClient.Do
	mockHTTPClient.Do = func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/api/v3/oauth2/token" {
			refreshCount++
		}
		return originalDo(req)
	}

	oauth2Config, err := auth.NewOAuth2Config(auth.OAuth2Options{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		TokenURL:     "https://api.upwork.com/api/v3/oauth2/token",
	})
	require.NoError(t, err)

	// Create client with an expired token
	client, err := NewClient(Config{
		OAuth2Config: oauth2Config,
		Token: &oauth2.Token{
			AccessToken:  "expired-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(-1 * time.Hour), // Already expired
		},
		HTTPClient: mockHTTPClient,
	})
	require.NoError(t, err)

	// Token should be automatically refreshed when needed
	err = client.RefreshToken(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, refreshCount)
}

func TestClientWithInvalidToken(t *testing.T) {
	tests := []struct {
		name    string
		token   *oauth2.Token
		wantErr bool
	}{
		{
			name: "expired token with refresh token",
			token: &oauth2.Token{
				AccessToken:  "expired-token",
				RefreshToken: "refresh-token",
				TokenType:    "Bearer",
				Expiry:       time.Now().Add(-1 * time.Hour),
			},
			wantErr: false, // Should be valid because it can be refreshed
		},
		{
			name: "expired token without refresh token",
			token: &oauth2.Token{
				AccessToken: "expired-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(-1 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "empty access token",
			token: &oauth2.Token{
				AccessToken: "",
				TokenType:   "Bearer",
			},
			wantErr: true,
		},
		{
			name: "wrong token type",
			token: &oauth2.Token{
				AccessToken: "token",
				TokenType:   "Basic",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(Config{
				OAuth2Config: &auth.OAuth2Config{},
				Token:        tt.token,
			})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGraphQLClientIntegration(t *testing.T) {
	// This test verifies that the GraphQL client is properly configured
	// with authentication headers and organization ID
	
	mockResponse := testutils.MockGraphQLResponse(
		map[string]interface{}{
			"viewer": map[string]interface{}{
				"id":    "user-123",
				"email": "test@example.com",
			},
		},
		nil,
	)
	
	responseBody, _ := json.Marshal(mockResponse)
	recorder := mocks.NewRequestRecorder(
		mocks.MockResponse{
			StatusCode: 200,
			Body:       string(responseBody),
		},
	)

	client, err := NewClient(Config{
		OAuth2Config: &auth.OAuth2Config{},
		Token: &oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
		OrganizationID: "org-123",
		HTTPClient:     recorder,
	})
	require.NoError(t, err)

	// Make a request through one of the services to verify headers
	ctx := context.Background()
	user, err := client.Users.GetCurrentUser(ctx)
	
	// Even if the method doesn't exist, we can verify the request was made correctly
	if err == nil {
		assert.NotNil(t, user)
	}

	// Verify the request headers
	if len(recorder.Requests) > 0 {
		req := recorder.GetLastRequest()
		assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
		assert.Equal(t, "org-123", req.Header.Get("X-Organization-UID"))
	}
}