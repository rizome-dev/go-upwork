// Package main provides a CLI tool for the Upwork SDK.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/rizome-dev/go-upwork/pkg"
)

func main() {
	// Define command-line flags
	clientID := flag.String("client-id", os.Getenv("UPWORK_CLIENT_ID"), "OAuth2 Client ID")
	clientSecret := flag.String("client-secret", os.Getenv("UPWORK_CLIENT_SECRET"), "OAuth2 Client Secret")
	orgID := flag.String("org-id", os.Getenv("UPWORK_ORG_ID"), "Organization ID")
	command := flag.String("cmd", "user", "Command to run (user, contracts, jobs)")

	flag.Parse()

	if *clientID == "" || *clientSecret == "" {
		fmt.Fprintln(os.Stderr, "Error: Client ID and Secret are required")
		fmt.Fprintln(os.Stderr, "Set UPWORK_CLIENT_ID and UPWORK_CLIENT_SECRET environment variables or use flags")
		os.Exit(1)
	}

	// Create client
	ctx := context.Background()
	config := &pkg.Config{
		ClientID:       *clientID,
		ClientSecret:   *clientSecret,
		OrganizationID: *orgID,
	}

	client, err := pkg.NewClient(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	// Execute command
	switch *command {
	case "user":
		user, err := client.Users.GetCurrentUser(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Current User: %s (%s %s)\n", user.Email, user.FirstName, user.LastName)

	case "contracts":
		// TODO: Implement contracts listing
		fmt.Println("Contracts command not yet implemented")

	case "jobs":
		// TODO: Implement jobs listing
		fmt.Println("Jobs command not yet implemented")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *command)
		os.Exit(1)
	}
}