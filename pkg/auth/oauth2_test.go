package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/rizome-dev/go-upwork/tests/mocks"
	"github.com/rizome-dev/go-upwork/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewOAuth2Config(t *testing.T) {
	tests := []struct {
		name     string
		opts     OAuth2Options
		wantErr  bool
		validate func(t *testing.T, cfg *OAuth2Config)
	}{
		{
			name: "valid configuration",
			opts: OAuth2Options{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost:8080/callback",
				Scopes:       []string{"read", "write"},
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *OAuth2Config) {
				assert.Equal(t, "test-client-id", cfg.config.ClientID)
				assert.Equal(t, "test-client-secret", cfg.config.ClientSecret)
				assert.Equal(t, "http://localhost:8080/callback", cfg.config.RedirectURL)
				assert.Equal(t, []string{"read", "write"}, cfg.config.Scopes)
			},
		},
		{
			name: "missing client ID",
			opts: OAuth2Options{
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost:8080/callback",
			},
			wantErr: true,
		},
		{
			name: "missing client secret",
			opts: OAuth2Options{
				ClientID:    "test-client-id",
				RedirectURL: "http://localhost:8080/callback",
			},
			wantErr: true,
		},
		{
			name: "default scopes",
			opts: OAuth2Options{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost:8080/callback",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *OAuth2Config) {
				expectedScopes := []string{
					"openid", "email", "profile", "offline_access",
					"contracts:read", "contracts:write",
					"messages:read", "messages:write",
					"jobs:read", "jobs:write",
					"proposals:read", "proposals:write",
					"organizations:read",
				}
				assert.Equal(t, expectedScopes, cfg.config.Scopes)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewOAuth2Config(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestGetAuthorizationURL(t *testing.T) {
	cfg, err := NewOAuth2Config(OAuth2Options{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read", "write"},
	})
	require.NoError(t, err)

	authURL := cfg.GetAuthorizationURL("test-state")
	
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)

	assert.Equal(t, "https", parsedURL.Scheme)
	assert.Equal(t, "www.upwork.com", parsedURL.Host)
	assert.Equal(t, "/ab/account-security/oauth2/authorize", parsedURL.Path)
	
	query := parsedURL.Query()
	assert.Equal(t, "test-client-id", query.Get("client_id"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "http://localhost:8080/callback", query.Get("redirect_uri"))
	assert.Equal(t, "read write", query.Get("scope"))
	assert.Equal(t, "test-state", query.Get("state"))
}

func TestExchangeAuthorizationCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/oauth2/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		
		err := r.ParseForm()
		require.NoError(t, err)
		
		assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
		assert.Equal(t, "test-code", r.Form.Get("code"))
		assert.Equal(t, "http://localhost:8080/callback", r.Form.Get("redirect_uri"))
		
		// Check basic auth
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "test-client-id", username)
		assert.Equal(t, "test-client-secret", password)
		
		response := testutils.MockOAuth2Token("test-access-token", 3600)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg, err := NewOAuth2Config(OAuth2Options{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		AuthURL:      server.URL + "/ab/account-security/oauth2/authorize",
		TokenURL:     server.URL + "/api/v3/oauth2/token",
	})
	require.NoError(t, err)

	token, err := cfg.ExchangeAuthorizationCode(context.Background(), "test-code")
	require.NoError(t, err)
	
	assert.Equal(t, "test-access-token", token.AccessToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.Equal(t, "mock_refresh_token", token.RefreshToken)
	assert.True(t, token.Valid())
}

func TestRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/oauth2/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		
		err := r.ParseForm()
		require.NoError(t, err)
		
		assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
		assert.Equal(t, "old-refresh-token", r.Form.Get("refresh_token"))
		
		response := testutils.MockOAuth2Token("new-access-token", 3600)
		response["refresh_token"] = "new-refresh-token"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg, err := NewOAuth2Config(OAuth2Options{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		TokenURL:     server.URL + "/api/v3/oauth2/token",
	})
	require.NoError(t, err)

	oldToken := &oauth2.Token{
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // Expired
	}

	newToken, err := cfg.RefreshToken(context.Background(), oldToken)
	require.NoError(t, err)
	
	assert.Equal(t, "new-access-token", newToken.AccessToken)
	assert.Equal(t, "new-refresh-token", newToken.RefreshToken)
	assert.True(t, newToken.Valid())
}

func TestClientCredentialsGrant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/oauth2/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		
		err := r.ParseForm()
		require.NoError(t, err)
		
		assert.Equal(t, "client_credentials", r.Form.Get("grant_type"))
		assert.Equal(t, "read write", r.Form.Get("scope"))
		
		// Check basic auth
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "test-client-id", username)
		assert.Equal(t, "test-client-secret", password)
		
		response := testutils.MockOAuth2Token("client-access-token", 3600)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg, err := NewOAuth2Config(OAuth2Options{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenURL:     server.URL + "/api/v3/oauth2/token",
		Scopes:       []string{"read", "write"},
	})
	require.NoError(t, err)

	token, err := cfg.ClientCredentialsGrant(context.Background())
	require.NoError(t, err)
	
	assert.Equal(t, "client-access-token", token.AccessToken)
	assert.True(t, token.Valid())
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name    string
		token   *oauth2.Token
		wantErr bool
	}{
		{
			name: "valid token",
			token: &oauth2.Token{
				AccessToken: "valid-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(1 * time.Hour),
			},
			wantErr: false,
		},
		{
			name:    "nil token",
			token:   nil,
			wantErr: true,
		},
		{
			name: "empty access token",
			token: &oauth2.Token{
				AccessToken: "",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(1 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "expired token",
			token: &oauth2.Token{
				AccessToken: "expired-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(-1 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "wrong token type",
			token: &oauth2.Token{
				AccessToken: "valid-token",
				TokenType:   "Basic",
				Expiry:      time.Now().Add(1 * time.Hour),
			},
			wantErr: true,
		},
	}

	cfg, _ := NewOAuth2Config(OAuth2Options{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cfg.ValidateToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenExpiry(t *testing.T) {
	cfg, _ := NewOAuth2Config(OAuth2Options{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	tests := []struct {
		name           string
		token          *oauth2.Token
		expectedExpiry time.Duration
		expectError    bool
	}{
		{
			name: "token with future expiry",
			token: &oauth2.Token{
				Expiry: time.Now().Add(30 * time.Minute),
			},
			expectedExpiry: 30 * time.Minute,
			expectError:    false,
		},
		{
			name: "expired token",
			token: &oauth2.Token{
				Expiry: time.Now().Add(-30 * time.Minute),
			},
			expectedExpiry: -30 * time.Minute,
			expectError:    false,
		},
		{
			name:        "nil token",
			token:       nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiry, err := cfg.TokenExpiry(tt.token)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Allow for small time differences during test execution
				assert.InDelta(t, tt.expectedExpiry.Seconds(), expiry.Seconds(), 1)
			}
		})
	}
}

func TestHTTPClient(t *testing.T) {
	// Test with custom HTTP client
	customClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	cfg, err := NewOAuth2Config(OAuth2Options{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		HTTPClient:   customClient,
	})
	require.NoError(t, err)

	token := &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	httpClient := cfg.HTTPClient(context.Background(), token)
	assert.NotNil(t, httpClient)
}

func TestOAuth2ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		responseCode int
		responseBody string
		expectError  string
	}{
		{
			name:         "invalid grant error",
			responseCode: 400,
			responseBody: `{"error":"invalid_grant","error_description":"Invalid authorization code"}`,
			expectError:  "invalid_grant",
		},
		{
			name:         "invalid client error",
			responseCode: 401,
			responseBody: `{"error":"invalid_client","error_description":"Invalid client credentials"}`,
			expectError:  "invalid_client",
		},
		{
			name:         "server error",
			responseCode: 500,
			responseBody: `{"error":"server_error","error_description":"Internal server error"}`,
			expectError:  "server_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			cfg, err := NewOAuth2Config(OAuth2Options{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost:8080/callback",
				TokenURL:     server.URL + "/api/v3/oauth2/token",
			})
			require.NoError(t, err)

			_, err = cfg.ExchangeAuthorizationCode(context.Background(), "test-code")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestOAuth2Scopes(t *testing.T) {
	tests := []struct {
		name           string
		inputScopes    []string
		expectedScopes []string
	}{
		{
			name:        "custom scopes",
			inputScopes: []string{"custom:read", "custom:write"},
			expectedScopes: []string{"custom:read", "custom:write"},
		},
		{
			name:        "empty scopes uses defaults",
			inputScopes: nil,
			expectedScopes: []string{
				"openid", "email", "profile", "offline_access",
				"contracts:read", "contracts:write",
				"messages:read", "messages:write",
				"jobs:read", "jobs:write",
				"proposals:read", "proposals:write",
				"organizations:read",
			},
		},
		{
			name:        "duplicate scopes are preserved",
			inputScopes: []string{"read", "write", "read"},
			expectedScopes: []string{"read", "write", "read"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewOAuth2Config(OAuth2Options{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Scopes:       tt.inputScopes,
			})
			require.NoError(t, err)
			
			assert.Equal(t, tt.expectedScopes, cfg.config.Scopes)
		})
	}
}

func TestConcurrentTokenRefresh(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		
		response := testutils.MockOAuth2Token("refreshed-token", 3600)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg, err := NewOAuth2Config(OAuth2Options{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenURL:     server.URL + "/api/v3/oauth2/token",
	})
	require.NoError(t, err)

	expiredToken := &oauth2.Token{
		AccessToken:  "old-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	// Launch multiple concurrent refresh attempts
	const numGoroutines = 5
	results := make(chan *oauth2.Token, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			token, err := cfg.RefreshToken(context.Background(), expiredToken)
			if err != nil {
				errors <- err
			} else {
				results <- token
			}
		}()
	}

	// Collect results
	var tokens []*oauth2.Token
	for i := 0; i < numGoroutines; i++ {
		select {
		case token := <-results:
			tokens = append(tokens, token)
		case err := <-errors:
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	// All tokens should be valid
	for _, token := range tokens {
		assert.Equal(t, "refreshed-token", token.AccessToken)
		assert.True(t, token.Valid())
	}
}

func TestTokenSourceWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a delay to test context cancellation
		select {
		case <-time.After(1 * time.Second):
			response := testutils.MockOAuth2Token("new-token", 3600)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	cfg, err := NewOAuth2Config(OAuth2Options{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenURL:     server.URL + "/api/v3/oauth2/token",
	})
	require.NoError(t, err)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = cfg.ClientCredentialsGrant(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}