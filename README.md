# Upwork Go SDK

[![GoDoc](https://pkg.go.dev/badge/github.com/rizome-dev/go-upwork)](https://pkg.go.dev/github.com/rizome-dev/go-upwork)
[![Go Report Card](https://goreportcard.com/badge/github.com/rizome-dev/go-upwork)](https://goreportcard.com/report/github.com/rizome-dev/go-upwork)

```shell
go get github.com/rizome-dev/go-upwork
```

built by: [rizome labs](https://rizome.dev)

contact us: [hi (at) rizome.dev](mailto:hi@rizome.dev)

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/rizome-dev/go-upwork/pkg"
)

func main() {
    // Create client configuration
    config := &pkg.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "your-redirect-url",
    }
    
    // Create client
    ctx := context.Background()
    client, err := pkg.NewClient(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get authorization URL
    authURL := client.GetAuthURL("state")
    fmt.Println("Visit:", authURL)
    
    // Exchange authorization code for token
    token, err := client.ExchangeCode(ctx, "auth-code")
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the authenticated client
    user, err := client.Users.GetCurrentUser(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Hello, %s!\n", user.Name)
}
```

## Core Services

### Authentication

```go
// OAuth2 Authorization Code Flow
authURL := client.GetAuthURL("state")
token, err := client.ExchangeCode(ctx, "code")

// Refresh token
newToken, err := client.RefreshToken(ctx)

// Service Account (Enterprise)
config.ServiceAccount = true
```

### Users & Organizations

```go
// Get current user
user, err := client.Users.GetCurrentUser(ctx)

// Get user by ID
user, err := client.Users.GetUserByID(ctx, "user-id")

// Get organization
org, err := client.Users.GetOrganization(ctx)

// List companies
companies, err := client.Users.GetCompanySelector(ctx)
```

### Contracts & Milestones

```go
// List contracts
contracts, err := client.Contracts.ListContracts(ctx, api.ListContractsInput{
    Filter: &api.ContractFilter{
        Status: []api.ContractStatus{api.ContractStatusActive},
    },
})

// Create milestone
milestone, err := client.Contracts.CreateMilestone(ctx, api.CreateMilestoneInput{
    ContractID:    "contract-id",
    Description:   "Milestone 1",
    DepositAmount: "1000.00",
    DueDate:       "2024-12-31",
})

// End contract
err = client.Contracts.EndContractAsClient(ctx, api.EndContractInput{
    ContractID: "contract-id",
    Reason:     "Work completed",
})
```

### Job Postings

```go
// Search jobs
jobs, err := client.Jobs.SearchJobs(ctx, api.MarketplaceJobFilter{
    SearchExpression: "golang developer",
    JobType:         api.ContractTypeHourly,
    DaysPosted:      7,
})

// Create job posting
job, err := client.Jobs.CreateJobPosting(ctx, api.CreateJobPostingInput{
    Title:        "Go Developer Needed",
    Description:  "Looking for experienced Go developer",
    CategoryID:   "category-id",
    Skills:       []string{"golang", "api"},
    ContractType: api.ContractTypeHourly,
})
```

### Messaging

```go
// List rooms
rooms, err := client.Messages.ListRooms(ctx, &api.RoomFilter{
    UnreadRoomsOnly: true,
}, nil, api.SortOrderDesc)

// Send message
story, err := client.Messages.SendMessage(ctx, api.CreateStoryInput{
    RoomID:  "room-id",
    Message: "Hello!",
})

// Create room
room, err := client.Messages.CreateRoom(ctx, api.CreateRoomInput{
    RoomName: "Project Discussion",
    RoomType: api.RoomTypeGroup,
    Users: []api.RoomUserInput{
        {UserID: "user1", OrganizationID: "org1"},
        {UserID: "user2", OrganizationID: "org1"},
    },
})
```

### Reports & Analytics

```go
// Get time reports
reports, err := client.Reports.GetTimeReport(ctx, api.TimeReportInput{
    OrganizationID: "org-id",
    DateRange: api.DateRange{
        Start: time.Now().AddDate(0, -1, 0),
        End:   time.Now(),
    },
})

// Get work diary
diary, err := client.Reports.GetWorkDiaryByCompany(ctx, "company-id", "2024-01-15")
```

### Freelancer Profiles

```go
// Search freelancers
results, err := client.Freelancers.SearchFreelancers(ctx, api.SearchFreelancersInput{
    Skills:     []string{"golang", "python"},
    Countries:  []string{"US", "CA"},
    TopRated:   true,
    HourlyRate: &api.RangeFilter{Min: 50, Max: 150},
})

// Get profile
profile, err := client.Freelancers.GetFreelancerProfile(ctx, "profile-key")
```

## Advanced Features

### Concurrent Operations

```go
// Batch operations using goroutines
var wg sync.WaitGroup
errors := make(chan error, 3)

wg.Add(3)
go func() {
    defer wg.Done()
    if _, err := client.Users.GetCurrentUser(ctx); err != nil {
        errors <- err
    }
}()

go func() {
    defer wg.Done()
    if _, err := client.Contracts.ListContracts(ctx, input); err != nil {
        errors <- err
    }
}()

go func() {
    defer wg.Done()
    if _, err := client.Messages.ListRooms(ctx, nil, nil, ""); err != nil {
        errors <- err
    }
}()

wg.Wait()
close(errors)
```

### Error Handling

```go
// Check for specific error types
if err != nil {
    var apiErr *errors.APIError
    if errors.As(err, &apiErr) {
        if apiErr.IsRateLimited() {
            // Handle rate limiting
        } else if apiErr.IsUnauthorized() {
            // Refresh token
        }
    }
}
```

### Custom HTTP Client

```go
// Use custom HTTP client with proxy
httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    },
}

config := &pkg.Config{
    HTTPClient: httpClient,
    // ... other config
}
```

## Configuration

### Environment Variables

```bash
export UPWORK_CLIENT_ID=your-client-id
export UPWORK_CLIENT_SECRET=your-client-secret
export UPWORK_REDIRECT_URL=http://localhost:8080/callback
```

### Organization Context

```go
// Set default organization
client.SetOrganizationID("org-id")

// Or use per-request context
ctx = context.WithValue(ctx, "org-id", "different-org-id")
```

## Project Structure

```
go-upwork/
├── pkg/                  # Public API package
│   ├── client.go         # Main client implementation
│   ├── auth/             # OAuth2 authentication
│   ├── errors/           # Error types and handling
│   ├── models/           # Shared data models
│   └── services/         # API service implementations
├── internal/             # Internal packages
│   ├── graphql/          # GraphQL client internals
│   └── ratelimit/        # Rate limiting implementation
├── cmd/upwork-cli/       # CLI tool
├── examples/             # Usage examples
└── docs/                 # Additional documentation
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## Examples

See the [examples](examples/) directory for more detailed usage examples:

- [Basic Usage](examples/basic_usage.go) - Authentication and basic operations
- [Concurrent Operations](examples/concurrent.go) - Parallel API calls
- [Job Management](examples/jobs.go) - Creating and managing job postings
- [Messaging](examples/messaging.go) - Room and message management
- [Reports](examples/reports.go) - Analytics and reporting

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [API Documentation](https://www.upwork.com/developer/documentation/graphql/api/docs/index.html)
- 💬 [Stack Overflow](https://stackoverflow.com/questions/tagged/upwork-api)
- 🐛 [Issue Tracker](https://github.com/rizome-dev/go-upwork/issues)

## Acknowledgments

- Built with concurrent design patterns for optimal performance
- Follows Go best practices and idioms
- Comprehensive GraphQL support for modern API interactions
