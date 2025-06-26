package services

import (
	"context"
	"fmt"
	
	"github.com/rizome-dev/go-upwork/pkg/models"
)

// UsersService handles user-related API operations
type UsersService struct {
	client *BaseClient
}

// NewUsersService creates a new users service
func NewUsersService(client *BaseClient) *UsersService {
	return &UsersService{client: client}
}

// User represents a user in the Upwork system
type User struct {
	ID        models.ID       `json:"id"`
	Nid       string   `json:"nid"`
	Rid       string   `json:"rid"`
	Name      string   `json:"name"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Email     string   `json:"email"`
	PhotoURL  string   `json:"photoUrl"`
	PublicURL string   `json:"publicUrl"`
	Location  models.Location `json:"location"`
}


// GetCurrentUser returns the current authenticated user
func (s *UsersService) GetCurrentUser(ctx context.Context) (*User, error) {
	query := `
		query GetCurrentUser {
			user {
				id
				nid
				rid
				name
				firstName
				lastName
				email
				photoUrl
				publicUrl
				location {
					country
					state
					city
					timezone
					offsetToUTC
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
	}
	
	var resp struct {
		User User `json:"user"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.User, nil
}

// GetUserByID returns a user by their ID
func (s *UsersService) GetUserByID(ctx context.Context, userID string) (*User, error) {
	query := `
		query GetUserDetails($id: ID!) {
			userDetails(id: $id) {
				id
				nid
				rid
				name
				firstName
				lastName
				email
				photoUrl
				publicUrl
				location {
					country
					state
					city
					timezone
					offsetToUTC
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"id": userID,
		},
	}
	
	var resp struct {
		UserDetails User `json:"userDetails"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.UserDetails, nil
}

// GetUsersByEmail returns users matching the given email addresses
func (s *UsersService) GetUsersByEmail(ctx context.Context, emails []string) ([]User, error) {
	query := `
		query GetUsersByEmail($emails: [String!]!) {
			userIdsByEmail(emails: $emails) {
				email
				userId
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"emails": emails,
		},
	}
	
	var resp struct {
		UserIdsByEmail []struct {
			Email  string `json:"email"`
			UserID string `json:"userId"`
		} `json:"userIdsByEmail"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	// Get full user details for each ID
	users := make([]User, 0, len(resp.UserIdsByEmail))
	for _, emailUser := range resp.UserIdsByEmail {
		user, err := s.GetUserByID(ctx, emailUser.UserID)
		if err != nil {
			// Continue on error, don't fail entire request
			continue
		}
		users = append(users, *user)
	}
	
	return users, nil
}

// Company represents a company
type Company struct {
	ID          ID     `json:"id"`
	Name        string `json:"name"`
	CompanyName string `json:"companyName"`
}

// CompanySelector represents a company in the selector
type CompanySelector struct {
	Title          string `json:"title"`
	OrganizationID string `json:"organizationId"`
}

// GetCompanySelector returns the list of companies the user has access to
func (s *UsersService) GetCompanySelector(ctx context.Context) ([]CompanySelector, error) {
	query := `
		query GetCompanySelector {
			companySelector {
				items {
					title
					organizationId
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
	}
	
	var resp struct {
		CompanySelector struct {
			Items []CompanySelector `json:"items"`
		} `json:"companySelector"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return resp.CompanySelector.Items, nil
}

// Organization represents an organization
type Organization struct {
	ID                ID                  `json:"id"`
	Name              string              `json:"name"`
	Company           Company             `json:"company"`
	ChildOrganizations []Organization     `json:"childOrganizations"`
	ParentOrganization *Organization      `json:"parentOrganization"`
	Staff             []Staff            `json:"staff"`
}

// Staff represents a staff member
type Staff struct {
	User             User   `json:"user"`
	StaffType        string `json:"staffType"`
	ActivationStatus string `json:"activationStatus"`
}

// GetOrganization returns the current organization
func (s *UsersService) GetOrganization(ctx context.Context) (*Organization, error) {
	query := `
		query GetOrganization {
			organization {
				id
				name
				company {
					id
					name
					companyName
				}
				childOrganizations {
					id
					name
					company {
						id
						name
					}
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
	}
	
	var resp struct {
		Organization Organization `json:"organization"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.Organization, nil
}

// GetOrganizationStaff returns staff members for a child organization
func (s *UsersService) GetOrganizationStaff(ctx context.Context, childOrgID string) ([]Staff, error) {
	query := `
		query GetChildOrganizationStaff($childOrganizationId: ID!) {
			organization {
				childOrganization(id: $childOrganizationId) {
					staffs {
						edges {
							node {
								user {
									id
									name
									publicUrl
								}
								staffType
								activationStatus
							}
						}
					}
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"childOrganizationId": childOrgID,
		},
	}
	
	var resp struct {
		Organization struct {
			ChildOrganization struct {
				Staffs struct {
					Edges []struct {
						Node Staff `json:"node"`
					} `json:"edges"`
				} `json:"staffs"`
			} `json:"childOrganization"`
		} `json:"organization"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	staff := make([]Staff, 0, len(resp.Organization.ChildOrganization.Staffs.Edges))
	for _, edge := range resp.Organization.ChildOrganization.Staffs.Edges {
		staff = append(staff, edge.Node)
	}
	
	return staff, nil
}

// InviteToTeamInput represents input for inviting to team
type InviteToTeamInput struct {
	TeamID    string   `json:"teamId"`
	Emails    []string `json:"emails"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Message   string   `json:"message"`
}

// InviteToTeam invites users to a team
func (s *UsersService) InviteToTeam(ctx context.Context, input InviteToTeamInput) error {
	mutation := `
		mutation InviteToTeam($input: InviteToTeamInput!) {
			inviteToTeam(input: $input) {
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"input": input,
		},
	}
	
	var resp struct {
		InviteToTeam struct {
			Success bool `json:"success"`
		} `json:"inviteToTeam"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.InviteToTeam.Success {
		return fmt.Errorf("failed to invite to team")
	}
	
	return nil
}