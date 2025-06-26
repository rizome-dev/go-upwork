// Package auth provides OAuth2 authentication for the Upwork API.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	
	"golang.org/x/oauth2"
)

const (
	// AuthorizationURL is the OAuth2 authorization endpoint
	AuthorizationURL = "https://www.upwork.com/ab/account-security/oauth2/authorize"
	
	// TokenURL is the OAuth2 token endpoint
	TokenURL = "https://www.upwork.com/api/v3/oauth2/token"
)

// GrantType represents OAuth2 grant types
type GrantType string

const (
	// GrantTypeAuthorizationCode is the authorization code grant
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	
	// GrantTypeImplicit is the implicit grant
	GrantTypeImplicit GrantType = "token"
	
	// GrantTypeClientCredentials is the client credentials grant (enterprise only)
	GrantTypeClientCredentials GrantType = "client_credentials"
	
	// GrantTypeRefreshToken is the refresh token grant
	GrantTypeRefreshToken GrantType = "refresh_token"
)

// Scope represents an OAuth2 scope
type Scope string

// Common scopes
const (
	// Read-only scopes
	ScopeMessagingRead         Scope = "messages:read"
	ScopeContractsRead         Scope = "contracts:read"
	ScopeProfileRead           Scope = "profile:read"
	ScopeJobsRead              Scope = "jobs:read"
	ScopeReportsRead           Scope = "reports:read"
	ScopeActivitiesRead        Scope = "activities:read"
	ScopeMetadataRead          Scope = "metadata:read"
	ScopeOrganizationRead      Scope = "organization:read"
	ScopeTimesheetRead         Scope = "timesheet:read"
	ScopeSnapshotsRead         Scope = "snapshots:read"
	
	// Read-write scopes
	ScopeMessagingWrite        Scope = "messages:write"
	ScopeContractsWrite        Scope = "contracts:write"
	ScopeProfileWrite          Scope = "profile:write"
	ScopeJobsWrite             Scope = "jobs:write"
	ScopeActivitiesWrite       Scope = "activities:write"
	ScopePaymentsWrite         Scope = "payments:write"
	ScopeOffersWrite           Scope = "offers:write"
)

// GetDefaultScopes returns the default set of scopes
func GetDefaultScopes() []string {
	return []string{
		string(ScopeMessagingRead),
		string(ScopeContractsRead),
		string(ScopeProfileRead),
		string(ScopeJobsRead),
		string(ScopeOrganizationRead),
	}
}

// GetAllScopes returns all available scopes
func GetAllScopes() []string {
	return []string{
		string(ScopeMessagingRead),
		string(ScopeMessagingWrite),
		string(ScopeContractsRead),
		string(ScopeContractsWrite),
		string(ScopeProfileRead),
		string(ScopeProfileWrite),
		string(ScopeJobsRead),
		string(ScopeJobsWrite),
		string(ScopeReportsRead),
		string(ScopeActivitiesRead),
		string(ScopeActivitiesWrite),
		string(ScopeMetadataRead),
		string(ScopeOrganizationRead),
		string(ScopeTimesheetRead),
		string(ScopeSnapshotsRead),
		string(ScopePaymentsWrite),
		string(ScopeOffersWrite),
	}
}

// Config holds OAuth2 configuration
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	GrantType    GrantType
}

// Client handles OAuth2 authentication
type Client struct {
	config       *Config
	oauth2Config *oauth2.Config
	httpClient   *http.Client
}

// NewClient creates a new OAuth2 client
func NewClient(config *Config) *Client {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthorizationURL,
			TokenURL: TokenURL,
		},
	}
	
	return &Client{
		config:       config,
		oauth2Config: oauth2Config,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// GetAuthorizationURL returns the OAuth2 authorization URL
func (c *Client) GetAuthorizationURL(state string) string {
	params := url.Values{}
	params.Set("response_type", string(c.config.GrantType))
	params.Set("client_id", c.config.ClientID)
	params.Set("redirect_uri", c.config.RedirectURL)
	params.Set("state", state)
	
	if len(c.config.Scopes) > 0 {
		params.Set("scope", strings.Join(c.config.Scopes, " "))
	}
	
	return fmt.Sprintf("%s?%s", AuthorizationURL, params.Encode())
}

// ExchangeCode exchanges an authorization code for tokens
func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.oauth2Config.Exchange(ctx, code)
}

// RefreshToken refreshes an access token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Set("grant_type", string(GrantTypeRefreshToken))
	params.Set("client_id", c.config.ClientID)
	params.Set("client_secret", c.config.ClientSecret)
	params.Set("refresh_token", refreshToken)
	
	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}
	
	var token oauth2.Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	
	return &token, nil
}

// ClientCredentials obtains a token using client credentials grant
func (c *Client) ClientCredentials(ctx context.Context) (*oauth2.Token, error) {
	if c.config.GrantType != GrantTypeClientCredentials {
		return nil, fmt.Errorf("client credentials grant not configured")
	}
	
	params := url.Values{}
	params.Set("grant_type", string(GrantTypeClientCredentials))
	params.Set("client_id", c.config.ClientID)
	params.Set("client_secret", c.config.ClientSecret)
	
	if len(c.config.Scopes) > 0 {
		params.Set("scope", strings.Join(c.config.Scopes, " "))
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client credentials grant failed with status: %d", resp.StatusCode)
	}
	
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}
	
	token := &oauth2.Token{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		Expiry:      time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}
	
	return token, nil
}

// TokenSource returns an oauth2.TokenSource for the given token
func (c *Client) TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	return c.oauth2Config.TokenSource(ctx, token)
}

// HTTPClient returns an HTTP client with the OAuth2 token
func (c *Client) HTTPClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return c.oauth2Config.Client(ctx, token)
}

// ValidateToken checks if a token is valid and not expired
func ValidateToken(token *oauth2.Token) error {
	if token == nil {
		return fmt.Errorf("token is nil")
	}
	
	if token.AccessToken == "" {
		return fmt.Errorf("access token is empty")
	}
	
	if !token.Expiry.IsZero() && token.Expiry.Before(time.Now()) {
		return fmt.Errorf("token is expired")
	}
	
	return nil
}

// IsTokenExpired checks if a token is expired
func IsTokenExpired(token *oauth2.Token) bool {
	if token == nil || token.Expiry.IsZero() {
		return false
	}
	return token.Expiry.Before(time.Now())
}