package services

import (
	"github.com/rizome-dev/go-upwork/pkg/models"
	"context"
	"fmt"
)

// ContractsService handles contract-related API operations
type ContractsService struct {
	client *BaseClient
}

// NewContractsService creates a new contracts service
func NewContractsService(client *BaseClient) *ContractsService {
	return &ContractsService{client: client}
}

// Contract represents a contract
type Contract struct {
	ID                      models.ID                `json:"id"`
	Title                   string            `json:"title"`
	ContractType            ContractType      `json:"contractType"`
	Status                  ContractStatus    `json:"status"`
	CreatedDateTime         models.DateTime          `json:"createdDateTime"`
	StartDateTime           models.DateTime          `json:"startDateTime"`
	EndDateTime             *models.DateTime         `json:"endDateTime"`
	ModifiedDateTime        models.DateTime          `json:"modifiedDateTime"`
	HourlyChargeRate        *models.Money            `json:"hourlyChargeRate"`
	WeeklyHoursLimit        *int              `json:"weeklyHoursLimit"`
	WeeklyChargeAmount      *models.Money            `json:"weeklyChargeAmount"`
	ManualTimeAllowed       bool              `json:"manualTimeAllowed"`
	Paused                  bool              `json:"paused"`
	Suspended               bool              `json:"suspended"`
	Last                    bool              `json:"last"`
	Job                     *Job              `json:"job"`
	Offer                   *Offer            `json:"offer"`
	Freelancer              *FreelancerInfo   `json:"freelancer"`
	Client                  *ClientInfo       `json:"client"`
	Milestones              []Milestone       `json:"milestones"`
}

// ContractType represents the type of contract
type ContractType string

const (
	ContractTypeHourly     ContractType = "HOURLY"
	ContractTypeFixedPrice ContractType = "FIXED_PRICE"
)

// ContractStatus represents the status of a contract
type ContractStatus string

const (
	ContractStatusActive    ContractStatus = "ACTIVE"
	ContractStatusPaused    ContractStatus = "PAUSED"
	ContractStatusEnded     ContractStatus = "ENDED"
	ContractStatusSuspended ContractStatus = "SUSPENDED"
)

// Job represents a job
type Job struct {
	ID      ID         `json:"id"`
	Content JobContent `json:"content"`
}

// JobContent represents job content
type JobContent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Offer represents an offer
type Offer struct {
	ID         ID         `json:"id"`
	OfferTerms OfferTerms `json:"offerTerms"`
}

// OfferTerms represents offer terms
type OfferTerms struct {
	HourlyTerm     *HourlyTerm     `json:"hourlyTerm"`
	FixedPriceTerm *FixedPriceTerm `json:"fixedPriceTerm"`
}

// HourlyTerm represents hourly contract terms
type HourlyTerm struct {
	HourlyRate       Money `json:"hourlyRate"`
	WeeklyHoursLimit int   `json:"weeklyHoursLimit"`
}

// FixedPriceTerm represents fixed price contract terms
type FixedPriceTerm struct {
	Budget Money `json:"budget"`
}

// FreelancerInfo represents freelancer information
type FreelancerInfo struct {
	User            User            `json:"user"`
	CountryDetails  CountryDetails  `json:"countryDetails"`
}

// ClientInfo represents client information
type ClientInfo struct {
	User User `json:"user"`
}

// CountryDetails represents country information
type CountryDetails struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
}

// GetContract returns a contract by ID
func (s *ContractsService) GetContract(ctx context.Context, contractID string) (*Contract, error) {
	query := `
		query GetContract($id: ID!) {
			contract(id: $id) {
				id
				title
				contractType
				status
				createdDateTime
				startDateTime
				endDateTime
				modifiedDateTime
				hourlyChargeRate {
					rawValue
					currency
					displayValue
				}
				weeklyHoursLimit
				weeklyChargeAmount {
					rawValue
					currency
					displayValue
				}
				manualTimeAllowed
				paused
				suspended
				last
				job {
					id
					content {
						title
						description
					}
				}
				offer {
					id
				}
				freelancer {
					user {
						id
						nid
						rid
						name
					}
					countryDetails {
						id
						name
					}
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"id": contractID,
		},
	}
	
	var resp struct {
		Contract Contract `json:"contract"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.Contract, nil
}

// ListContractsInput represents input for listing contracts
type ListContractsInput struct {
	Pagination *PaginationInput  `json:"pagination,omitempty"`
	Filter     *ContractFilter   `json:"filter,omitempty"`
}

// ContractFilter represents contract filtering options
type ContractFilter struct {
	Status       []ContractStatus `json:"status,omitempty"`
	ContractType []ContractType   `json:"contractType,omitempty"`
}

// ContractList represents a paginated list of contracts
type ContractList struct {
	TotalCount int        `json:"totalCount"`
	PageInfo   PageInfo   `json:"pageInfo"`
	Edges      []ContractEdge `json:"edges"`
}

// ContractEdge represents a contract edge in pagination
type ContractEdge struct {
	Cursor string   `json:"cursor"`
	Node   Contract `json:"node"`
}

// ListContracts returns a list of contracts
func (s *ContractsService) ListContracts(ctx context.Context, input ListContractsInput) (*ContractList, error) {
	query := `
		query ListContracts($pagination: Pagination, $filter: ContractFilter) {
			contractList(pagination: $pagination, filter: $filter) {
				totalCount
				pageInfo {
					hasNextPage
					hasPreviousPage
					startCursor
					endCursor
				}
				edges {
					cursor
					node {
						id
						title
						contractType
						status
						createdDateTime
						startDateTime
						hourlyChargeRate {
							rawValue
							currency
						}
						freelancer {
							user {
								id
								name
							}
						}
					}
				}
			}
		}
	`
	
	variables := map[string]interface{}{}
	if input.Pagination != nil {
		variables["pagination"] = input.Pagination
	}
	if input.Filter != nil {
		variables["filter"] = input.Filter
	}
	
	req := &GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	
	var resp struct {
		ContractList ContractList `json:"contractList"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.ContractList, nil
}

// EndContractInput represents input for ending a contract
type EndContractInput struct {
	ContractID     string  `json:"contractId"`
	Reason         string  `json:"reason"`
	Message        string  `json:"message"`
	Rating         *int    `json:"rating,omitempty"`
	Feedback       string  `json:"feedback,omitempty"`
}

// EndContractAsClient ends a contract from the client side
func (s *ContractsService) EndContractAsClient(ctx context.Context, input EndContractInput) error {
	mutation := `
		mutation EndContractByClient($input: EndContractByClientInput!) {
			endContractByClient(input: $input) {
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
		EndContractByClient struct {
			Success bool `json:"success"`
		} `json:"endContractByClient"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.EndContractByClient.Success {
		return fmt.Errorf("failed to end contract")
	}
	
	return nil
}

// EndContractAsFreelancer ends a contract from the freelancer side
func (s *ContractsService) EndContractAsFreelancer(ctx context.Context, input EndContractInput) error {
	mutation := `
		mutation EndContractByFreelancer($input: EndContractByFreelancerInput!) {
			endContractByFreelancer(input: $input) {
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
		EndContractByFreelancer struct {
			Success bool `json:"success"`
		} `json:"endContractByFreelancer"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.EndContractByFreelancer.Success {
		return fmt.Errorf("failed to end contract")
	}
	
	return nil
}

// PauseContract pauses a contract
func (s *ContractsService) PauseContract(ctx context.Context, contractID string) error {
	mutation := `
		mutation PauseContract($contractId: ID!) {
			pauseContract(contractId: $contractId) {
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"contractId": contractID,
		},
	}
	
	var resp struct {
		PauseContract struct {
			Success bool `json:"success"`
		} `json:"pauseContract"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.PauseContract.Success {
		return fmt.Errorf("failed to pause contract")
	}
	
	return nil
}

// RestartContract restarts a paused contract
func (s *ContractsService) RestartContract(ctx context.Context, contractID string) error {
	mutation := `
		mutation RestartContract($contractId: ID!) {
			restartContract(contractId: $contractId) {
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"contractId": contractID,
		},
	}
	
	var resp struct {
		RestartContract struct {
			Success bool `json:"success"`
		} `json:"restartContract"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.RestartContract.Success {
		return fmt.Errorf("failed to restart contract")
	}
	
	return nil
}

// UpdateHourlyLimitInput represents input for updating hourly limit
type UpdateHourlyLimitInput struct {
	ContractID       string `json:"contractId"`
	WeeklyHoursLimit int    `json:"weeklyHoursLimit"`
}

// UpdateContractHourlyLimit updates the weekly hours limit for a contract
func (s *ContractsService) UpdateContractHourlyLimit(ctx context.Context, input UpdateHourlyLimitInput) error {
	mutation := `
		mutation UpdateContractHourlyLimit($input: UpdateContractHourlyLimitInput!) {
			updateContractHourlyLimit(input: $input) {
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
		UpdateContractHourlyLimit struct {
			Success bool `json:"success"`
		} `json:"updateContractHourlyLimit"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.UpdateContractHourlyLimit.Success {
		return fmt.Errorf("failed to update hourly limit")
	}
	
	return nil
}