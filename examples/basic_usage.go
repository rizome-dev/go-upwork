// Package main demonstrates basic usage of the Upwork Go SDK
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	
	"github.com/rizome-dev/go-upwork/pkg"
)

func main() {
	// Create a new client configuration
	config := &pkg.Config{
		ClientID:     os.Getenv("UPWORK_CLIENT_ID"),
		ClientSecret: os.Getenv("UPWORK_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("UPWORK_REDIRECT_URL"),
	}
	
	// Create a new client
	ctx := context.Background()
	client, err := pkg.NewClient(ctx, config)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	
	// Example 1: OAuth2 Authentication Flow
	// Get the authorization URL
	authURL := client.GetAuthURL("state-token")
	fmt.Println("Visit this URL to authorize:", authURL)
	fmt.Print("Enter the authorization code: ")
	
	var code string
	fmt.Scanln(&code)
	
	// Exchange code for token
	token, err := client.ExchangeCode(ctx, code)
	if err != nil {
		log.Fatal("Failed to exchange code:", err)
	}
	
	fmt.Println("Successfully authenticated!")
	fmt.Printf("Access Token: %s\n", token.AccessToken)
	
	// Example 2: Get Current User
	currentUser, err := client.Users.GetCurrentUser(ctx)
	if err != nil {
		log.Fatal("Failed to get current user:", err)
	}
	
	fmt.Printf("\n=== Current User ===\n")
	fmt.Printf("Name: %s\n", currentUser.Name)
	fmt.Printf("Email: %s\n", currentUser.Email)
	fmt.Printf("ID: %s\n", currentUser.ID)
	
	// Example 3: Get Company Selector
	companies, err := client.Users.GetCompanySelector(ctx)
	if err != nil {
		log.Fatal("Failed to get companies:", err)
	}
	
	fmt.Printf("\n=== Available Companies ===\n")
	for _, company := range companies {
		fmt.Printf("- %s (ID: %s)\n", company.Title, company.OrganizationID)
	}
	
	// Set the organization ID for subsequent requests
	if len(companies) > 0 {
		client.SetOrganizationID(companies[0].OrganizationID)
		fmt.Printf("\nSet organization to: %s\n", companies[0].Title)
	}
	
	// Example 4: List Contracts
	contractsResp, err := client.Contracts.ListContracts(ctx, services.ListContractsInput{
		Pagination: &services.PaginationInput{
			First: 10,
		},
		Filter: &services.ContractFilter{
			Status: []services.ContractStatus{services.ContractStatusActive},
		},
	})
	if err != nil {
		log.Fatal("Failed to list contracts:", err)
	}
	
	fmt.Printf("\n=== Active Contracts ===\n")
	fmt.Printf("Total contracts: %d\n", contractsResp.TotalCount)
	for _, edge := range contractsResp.Edges {
		contract := edge.Node
		fmt.Printf("- %s (Type: %s, Status: %s)\n", 
			contract.Title, contract.ContractType, contract.Status)
	}
	
	// Example 5: Search for Jobs
	jobSearchResp, err := client.Jobs.SearchJobs(ctx, services.MarketplaceJobFilter{
		SearchExpression: "golang developer",
		JobType:         services.ContractTypeHourly,
		DaysPosted:      7,
		Pagination: &services.PaginationInput{
			First: 5,
		},
	})
	if err != nil {
		log.Fatal("Failed to search jobs:", err)
	}
	
	fmt.Printf("\n=== Job Search Results ===\n")
	fmt.Printf("Found %d jobs matching 'golang developer'\n", jobSearchResp.TotalCount)
	for _, edge := range jobSearchResp.Edges {
		job := edge.Node
		fmt.Printf("- %s\n", job.Title)
	}
	
	// Example 6: List Chat Rooms
	roomsResp, err := client.Messages.ListRooms(ctx, 
		&services.RoomFilter{
			UnreadRoomsOnly: true,
		},
		&services.PaginationInput{
			First: 10,
		},
		services.SortOrderDesc,
	)
	if err != nil {
		log.Fatal("Failed to list rooms:", err)
	}
	
	fmt.Printf("\n=== Unread Messages ===\n")
	fmt.Printf("Total rooms with unread messages: %d\n", roomsResp.TotalCount)
	for _, edge := range roomsResp.Edges {
		room := edge.Node
		fmt.Printf("- %s (%d unread)\n", room.RoomName, room.NumUnread)
	}
	
	// Example 7: Get Metadata - Skills
	skills, err := client.Metadata.GetSkills(ctx, 20, 0)
	if err != nil {
		log.Fatal("Failed to get skills:", err)
	}
	
	fmt.Printf("\n=== Popular Skills ===\n")
	for _, skill := range skills[:10] {
		fmt.Printf("- %s\n", skill.PreferredLabel)
	}
	
	// Example 8: Send a Message (if we have rooms)
	if len(roomsResp.Edges) > 0 {
		firstRoom := roomsResp.Edges[0].Node
		story, err := client.Messages.SendMessage(ctx, services.CreateStoryInput{
			RoomID:  string(firstRoom.ID),
			Message: "Hello from the Upwork Go SDK!",
		})
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		} else {
			fmt.Printf("\n=== Sent Message ===\n")
			fmt.Printf("Message ID: %s\n", story.ID)
			fmt.Printf("Sent at: %s\n", story.CreatedDateTime.DisplayValue)
		}
	}
	
	// Example 9: Create a Job Posting (commented out to avoid creating real jobs)
	/*
	newJob, err := client.Jobs.CreateJobPosting(ctx, services.CreateJobPostingInput{
		Title:           "Go Developer Needed for SDK Development",
		Description:     "Looking for an experienced Go developer to help build SDK features.",
		CategoryID:      "531770282580668419", // Web Development
		SubCategoryID:   "531770282589057033", // Web Development
		Skills:          []string{"golang", "api-development", "sdk"},
		ContractType:    services.ContractTypeHourly,
		HourlyBudgetMin: floatPtr(50),
		HourlyBudgetMax: floatPtr(100),
		Duration:        "1_3_months",
		Workload:        "part_time",
		TeamID:          "your-team-id",
	})
	if err != nil {
		log.Printf("Failed to create job: %v", err)
	} else {
		fmt.Printf("\n=== Created Job ===\n")
		fmt.Printf("Job ID: %s\n", newJob.ID)
		fmt.Printf("Status: %s\n", newJob.Info.Status)
	}
	*/
	
	// Example 10: Concurrent Operations
	fmt.Printf("\n=== Concurrent API Calls ===\n")
	
	// Use goroutines to make multiple API calls concurrently
	type result struct {
		name string
		err  error
		data interface{}
	}
	
	results := make(chan result, 3)
	
	// Fetch user details
	go func() {
		user, err := client.Users.GetCurrentUser(ctx)
		results <- result{name: "user", err: err, data: user}
	}()
	
	// Fetch organization
	go func() {
		org, err := client.Users.GetOrganization(ctx)
		results <- result{name: "organization", err: err, data: org}
	}()
	
	// Fetch categories
	go func() {
		categories, err := client.Metadata.GetCategories(ctx)
		results <- result{name: "categories", err: err, data: categories}
	}()
	
	// Collect results
	for i := 0; i < 3; i++ {
		res := <-results
		if res.err != nil {
			fmt.Printf("Error fetching %s: %v\n", res.name, res.err)
		} else {
			fmt.Printf("Successfully fetched %s\n", res.name)
		}
	}
	
	fmt.Println("\nExample completed successfully!")
}

// Helper function to create float64 pointer
func floatPtr(f float64) *float64 {
	return &f
}