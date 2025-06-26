package services

import (
	"github.com/rizome-dev/go-upwork/pkg/models"
	"context"
	"fmt"
)

// ActivitiesService handles activity-related API operations
type ActivitiesService struct {
	client *BaseClient
}

// NewActivitiesService creates a new activities service
func NewActivitiesService(client *BaseClient) *ActivitiesService {
	return &ActivitiesService{client: client}
}

// Activity represents a team activity
type Activity struct {
	RecordID    string `json:"recordId"`
	CompanyID   string `json:"companyId"`
	UserID      string `json:"userId"`
	Code        string `json:"code"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// ActivityList represents a list of activities
type ActivityList struct {
	TotalCount int            `json:"totalCount"`
	Page       PageFilter     `json:"page"`
	Edges      []ActivityEdge `json:"edges"`
}

// ActivityEdge represents an activity edge
type ActivityEdge struct {
	Node Activity `json:"node"`
}

// PageFilter represents page filter
type PageFilter struct {
	PageOffset int `json:"pageOffset"`
	PageSize   int `json:"pageSize"`
}

// ActivityFilter represents activity filter options
type ActivityFilter struct {
	ContractID string   `json:"contractId,omitempty"`
	Codes      []string `json:"codes,omitempty"`
}

// ListTeamActivitiesInput represents input for listing team activities
type ListTeamActivitiesInput struct {
	OrgID  string          `json:"orgId"`
	TeamID string          `json:"teamId,omitempty"`
	Filter *ActivityFilter `json:"filter,omitempty"`
	Page   *PageFilter     `json:"page,omitempty"`
}

// ListTeamActivities returns team activities
func (s *ActivitiesService) ListTeamActivities(ctx context.Context, input ListTeamActivitiesInput) (*ActivityList, error) {
	query := `
		query TeamActivities(
			$orgId: ID!,
			$teamId: ID,
			$filter: ActivityFilterInput,
			$page: PageFilterInput
		) {
			teamActivities(
				orgId: $orgId,
				teamId: $teamId,
				filter: $filter,
				page: $page
			) {
				totalCount
				edges {
					node {
						recordId
						companyId
						userId
						code
						description
						url
					}
				}
				page {
					pageOffset
					pageSize
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"orgId": input.OrgID,
	}
	if input.TeamID != "" {
		variables["teamId"] = input.TeamID
	}
	if input.Filter != nil {
		variables["filter"] = input.Filter
	}
	if input.Page != nil {
		variables["page"] = input.Page
	}
	
	req := &GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	
	var resp struct {
		TeamActivities ActivityList `json:"teamActivities"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.TeamActivities, nil
}

// TeamActivityInput represents input for team activity operations
type TeamActivityInput struct {
	Code         string   `json:"code"`
	Description  string   `json:"description"`
	URL          string   `json:"url,omitempty"`
	ContractIDs  []string `json:"contractIds,omitempty"`
	AllInCompany bool     `json:"allInCompany,omitempty"`
}

// AddTeamActivity creates a new team activity
func (s *ActivitiesService) AddTeamActivity(ctx context.Context, orgID string, teamID string, input TeamActivityInput) error {
	mutation := `
		mutation AddTeamActivity(
			$orgId: ID!,
			$teamId: ID!,
			$request: TeamActivityInput!
		) {
			addTeamActivity(
				orgId: $orgId,
				teamId: $teamId,
				request: $request
			) {
				id
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"orgId":   orgID,
			"teamId":  teamID,
			"request": input,
		},
	}
	
	var resp struct {
		AddTeamActivity struct {
			ID      string `json:"id"`
			Success bool   `json:"success"`
		} `json:"addTeamActivity"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.AddTeamActivity.Success {
		return fmt.Errorf("failed to add team activity")
	}
	
	return nil
}

// UpdateTeamActivity updates an existing team activity
func (s *ActivitiesService) UpdateTeamActivity(ctx context.Context, orgID string, teamID string, input TeamActivityInput) error {
	mutation := `
		mutation UpdateTeamActivity(
			$orgId: ID!,
			$teamId: ID!,
			$request: UpdateTeamActivityRequest!
		) {
			updateTeamActivity(
				orgId: $orgId,
				teamId: $teamId,
				request: $request
			) {
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"orgId":   orgID,
			"teamId":  teamID,
			"request": input,
		},
	}
	
	var resp struct {
		UpdateTeamActivity struct {
			Success bool `json:"success"`
		} `json:"updateTeamActivity"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.UpdateTeamActivity.Success {
		return fmt.Errorf("failed to update team activity")
	}
	
	return nil
}

// ArchiveTeamActivity archives team activities
func (s *ActivitiesService) ArchiveTeamActivity(ctx context.Context, orgID string, teamID string, codes []string) error {
	mutation := `
		mutation ArchiveTeamActivity(
			$orgId: ID!,
			$teamId: ID!,
			$codes: [String!]!
		) {
			archiveTeamActivity(
				orgId: $orgId,
				teamId: $teamId,
				codes: $codes
			) {
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"orgId":  orgID,
			"teamId": teamID,
			"codes":  codes,
		},
	}
	
	var resp struct {
		ArchiveTeamActivity struct {
			Success bool `json:"success"`
		} `json:"archiveTeamActivity"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.ArchiveTeamActivity.Success {
		return fmt.Errorf("failed to archive team activity")
	}
	
	return nil
}

// UnarchiveTeamActivity unarchives team activities
func (s *ActivitiesService) UnarchiveTeamActivity(ctx context.Context, orgID string, teamID string, codes []string) error {
	mutation := `
		mutation UnarchiveTeamActivity(
			$orgId: ID!,
			$teamId: ID!,
			$codes: [String!]!
		) {
			unarchiveTeamActivity(
				orgId: $orgId,
				teamId: $teamId,
				codes: $codes
			) {
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"orgId":  orgID,
			"teamId": teamID,
			"codes":  codes,
		},
	}
	
	var resp struct {
		UnarchiveTeamActivity struct {
			Success bool `json:"success"`
		} `json:"unarchiveTeamActivity"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.UnarchiveTeamActivity.Success {
		return fmt.Errorf("failed to unarchive team activity")
	}
	
	return nil
}

// AssignActivityToContract assigns activities to a contract
func (s *ActivitiesService) AssignActivityToContract(ctx context.Context, orgID string, teamID string, contractID string, codes []string) error {
	mutation := `
		mutation AssignTeamActivityToTheContract(
			$orgId: ID!,
			$teamId: ID!,
			$contractId: ID!,
			$codes: [String!]!
		) {
			assignTeamActivityToTheContract(
				orgId: $orgId,
				teamId: $teamId,
				contractId: $contractId,
				codes: $codes
			) {
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"orgId":      orgID,
			"teamId":     teamID,
			"contractId": contractID,
			"codes":      codes,
		},
	}
	
	var resp struct {
		AssignTeamActivityToTheContract struct {
			Success bool `json:"success"`
		} `json:"assignTeamActivityToTheContract"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.AssignTeamActivityToTheContract.Success {
		return fmt.Errorf("failed to assign activity to contract")
	}
	
	return nil
}