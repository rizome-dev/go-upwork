// Package pkg provides a Go client for the Upwork API.
package pkg

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/rizome-dev/go-upwork/internal/ratelimit"
	"github.com/rizome-dev/go-upwork/pkg/auth"
	"github.com/rizome-dev/go-upwork/pkg/errors"
	"github.com/rizome-dev/go-upwork/pkg/services"
	"golang.org/x/oauth2"
)

const (
	// DefaultAPIURL is the default Upwork API endpoint
	DefaultAPIURL = "https://api.upwork.com/graphql"
	
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
	
	// MaxRetries is the maximum number of retry attempts
	MaxRetries = 3
	
	// RateLimitPerMinute is the API rate limit
	RateLimitPerMinute = 300
)

// Client is the main Upwork API client
type Client struct {
	// HTTP client for making requests
	httpClient *http.Client
	
	// OAuth2 configuration
	oauth2Config *oauth2.Config
	
	// OAuth2 token
	token *oauth2.Token
	
	// API base URL
	apiURL string
	
	// Organization ID for X-Upwork-API-TenantId header
	organizationID string
	
	// Rate limiter
	rateLimiter *ratelimit.Limiter
	
	// Service clients
	Users       *services.UsersService
	Contracts   *services.ContractsService
	Jobs        *services.JobsService
	Messages    *services.MessagesService
	Freelancers *services.FreelancersService
	Reports     *services.ReportsService
	Activities  *services.ActivitiesService
	Metadata    *services.MetadataService
	
	// Base client for services
	baseClient *services.BaseClient
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Config holds configuration options for the client
type Config struct {
	// OAuth2 client credentials
	ClientID     string
	ClientSecret string
	RedirectURL  string
	
	// Optional: API endpoint URL (defaults to production)
	APIURL string
	
	// Optional: HTTP client (defaults to new client with timeout)
	HTTPClient *http.Client
	
	// Optional: Default organization ID
	OrganizationID string
	
	// Optional: OAuth2 token (for pre-authenticated clients)
	Token *oauth2.Token
	
	// Optional: Service account mode
	ServiceAccount bool
	
	// Optional: Custom scopes (defaults to GetDefaultScopes)
	Scopes []string
}

// NewClient creates a new Upwork API client
func NewClient(ctx context.Context, config *Config) (*Client, error) {
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, errors.ErrMissingCredentials
	}
	
	// Set defaults
	if config.APIURL == "" {
		config.APIURL = DefaultAPIURL
	}
	
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}
	
	if len(config.Scopes) == 0 {
		config.Scopes = auth.GetDefaultScopes()
	}
	
	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  auth.AuthorizationURL,
			TokenURL: auth.TokenURL,
		},
	}
	
	// Create rate limiter
	rl := ratelimit.New(RateLimitPerMinute, time.Minute)
	
	// Initialize client
	client := &Client{
		httpClient:     config.HTTPClient,
		oauth2Config:   oauth2Config,
		token:          config.Token,
		apiURL:         config.APIURL,
		organizationID: config.OrganizationID,
		rateLimiter:    rl,
	}
	
	// If token is provided, create OAuth2 client
	if config.Token != nil {
		client.httpClient = oauth2Config.Client(ctx, config.Token)
	}
	
	// Initialize services
	client.initServices()
	
	return client, nil
}

// SetOrganizationID sets the organization ID for API requests
func (c *Client) SetOrganizationID(orgID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.organizationID = orgID
	
	// Update base client
	if c.baseClient != nil {
		c.baseClient.OrganizationID = orgID
	}
}

// GetOrganizationID returns the current organization ID
func (c *Client) GetOrganizationID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.organizationID
}

// SetToken sets the OAuth2 token
func (c *Client) SetToken(ctx context.Context, token *oauth2.Token) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
	c.httpClient = c.oauth2Config.Client(ctx, token)
	c.initServices()
}

// GetToken returns the current OAuth2 token
func (c *Client) GetToken() *oauth2.Token {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

// GetAuthURL returns the OAuth2 authorization URL
func (c *Client) GetAuthURL(state string) string {
	return c.oauth2Config.AuthCodeURL(state)
}

// GetAuthURLWithOptions returns the OAuth2 authorization URL with additional options
func (c *Client) GetAuthURLWithOptions(state string, opts ...oauth2.AuthCodeOption) string {
	return c.oauth2Config.AuthCodeURL(state, opts...)
}

// ExchangeCode exchanges an authorization code for an access token
func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.WrapError(err, "failed to exchange authorization code")
	}
	
	c.SetToken(ctx, token)
	return token, nil
}

// RefreshToken refreshes the OAuth2 token
func (c *Client) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.token == nil || c.token.RefreshToken == "" {
		return nil, errors.ErrNoRefreshToken
	}
	
	tokenSource := c.oauth2Config.TokenSource(ctx, c.token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, errors.WrapError(err, "failed to refresh token")
	}
	
	c.token = newToken
	c.httpClient = c.oauth2Config.Client(ctx, newToken)
	c.initServices()
	
	return newToken, nil
}

// IsTokenExpired checks if the current token is expired
func (c *Client) IsTokenExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.token == nil {
		return true
	}
	
	return auth.IsTokenExpired(c.token)
}

// initServices initializes all service clients
func (c *Client) initServices() {
	c.baseClient = &services.BaseClient{
		HTTPClient:     c.httpClient,
		APIURL:         c.apiURL,
		OrganizationID: c.organizationID,
		RateLimiter:    c.rateLimiter,
	}
	
	c.Users = services.NewUsersService(c.baseClient)
	c.Contracts = services.NewContractsService(c.baseClient)
	c.Jobs = services.NewJobsService(c.baseClient)
	c.Messages = services.NewMessagesService(c.baseClient)
	c.Freelancers = services.NewFreelancersService(c.baseClient)
	c.Reports = services.NewReportsService(c.baseClient)
	c.Activities = services.NewActivitiesService(c.baseClient)
	c.Metadata = services.NewMetadataService(c.baseClient)
}