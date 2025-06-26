package services

import (
	"github.com/rizome-dev/go-upwork/pkg/models"
	"context"
)

// FreelancersService handles freelancer-related API operations
type FreelancersService struct {
	client *BaseClient
}

// NewFreelancersService creates a new freelancers service
func NewFreelancersService(client *BaseClient) *FreelancersService {
	return &FreelancersService{client: client}
}

// FreelancerProfile represents a freelancer profile
type FreelancerProfile struct {
	Identity      ProfileIdentity      `json:"identity"`
	PersonalData  PersonalData         `json:"personalData"`
	Aggregates    ProfileAggregates    `json:"aggregates"`
	Skills        []ProfileSkill       `json:"skills"`
	JobCategories []JobCategory        `json:"jobCategories"`
	Preferences   ProfilePreferences   `json:"preferences"`
}

// ProfileIdentity represents profile identity information
type ProfileIdentity struct {
	ID         ID     `json:"id"`
	Ciphertext string `json:"ciphertext"`
}

// PersonalData represents personal data
type PersonalData struct {
	FirstName   string      `json:"firstName"`
	LastName    string      `json:"lastName"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Portrait    Portrait    `json:"portrait"`
	Location    Location    `json:"location"`
}

// Portrait represents profile portrait URLs
type Portrait struct {
	Portrait    string `json:"portrait"`
	Portrait32  string `json:"portrait32"`
	Portrait50  string `json:"portrait50"`
	Portrait100 string `json:"portrait100"`
}

// ProfileAggregates represents profile aggregates
type ProfileAggregates struct {
	TotalHours              float64   `json:"totalHours"`
	TotalJobs               int       `json:"totalJobs"`
	TotalFeedback           int       `json:"totalFeedback"`
	AdjustedFeedbackScore   float64   `json:"adjustedFeedbackScore"`
	LastWorkedOn            *DateTime `json:"lastWorkedOn"`
	TopRatedStatus          bool      `json:"topRatedStatus"`
}

// ProfileSkill represents a skill in the profile
type ProfileSkill struct {
	Skill       Skill `json:"skill"`
	SkillUID    string `json:"skillUid"`
}

// JobCategory represents a job category
type JobCategory struct {
	ID     ID     `json:"id"`
	Name   string `json:"name"`
	Groups []CategoryGroup `json:"groups"`
}

// CategoryGroup represents a category group
type CategoryGroup struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
}

// ProfilePreferences represents profile preferences
type ProfilePreferences struct {
	VisibilityLevel string `json:"visibilityLevel"`
}

// GetFreelancerProfile retrieves a freelancer profile by profile key
func (s *FreelancersService) GetFreelancerProfile(ctx context.Context, profileKey string) (*FreelancerProfile, error) {
	query := `
		query GetFreelancerProfile($profileKey: String!) {
			freelancerProfileByProfileKey(profileKey: $profileKey) {
				identity {
					id
					ciphertext
				}
				personalData {
					firstName
					lastName
					title
					description
					portrait {
						portrait
						portrait32
						portrait50
						portrait100
					}
					location {
						country
						state
						city
						timezone
					}
				}
				aggregates {
					totalHours
					totalJobs
					totalFeedback
					adjustedFeedbackScore
					lastWorkedOn
					topRatedStatus
				}
				skills {
					skill {
						id
						prettyName
					}
					skillUid
				}
				jobCategories {
					id
					name
				}
				preferences {
					visibilityLevel
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"profileKey": profileKey,
		},
	}
	
	var resp struct {
		FreelancerProfileByProfileKey FreelancerProfile `json:"freelancerProfileByProfileKey"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.FreelancerProfileByProfileKey, nil
}

// SearchFreelancersInput represents input for searching freelancers
type SearchFreelancersInput struct {
	UserQuery      string           `json:"userQuery,omitempty"`
	Title          string           `json:"title,omitempty"`
	Skills         []string         `json:"skillsNames,omitempty"`
	Countries      []string         `json:"countries,omitempty"`
	HourlyRate     *RangeFilter     `json:"hourlyRate,omitempty"`
	JobSuccessScore *RangeFilter     `json:"jobSuccessScore,omitempty"`
	TotalJobs      *RangeFilter     `json:"totalJobs,omitempty"`
	TopRated       bool             `json:"topRated,omitempty"`
	Pagination     *PaginationInput `json:"paging,omitempty"`
}

// RangeFilter represents a range filter
type RangeFilter struct {
	Min float64 `json:"min,omitempty"`
	Max float64 `json:"max,omitempty"`
}

// FreelancerSearchResult represents freelancer search results
type FreelancerSearchResult struct {
	Profiles []FreelancerProfile `json:"profiles"`
}

// SearchFreelancers searches for freelancers
func (s *FreelancersService) SearchFreelancers(ctx context.Context, input SearchFreelancersInput) (*FreelancerSearchResult, error) {
	query := `
		query SearchFreelancers($request: FreelancerSearchRequest!) {
			search {
				searchFreelancerPublicProfile(request: $request) {
					profiles {
						profile {
							identity {
								id
								ciphertext
							}
							personalData {
								firstName
								lastName
								title
								location {
									country
									state
									city
								}
							}
							profileAggregates {
								totalJobs
								totalFeedback
								topRatedStatus
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
			"request": input,
		},
	}
	
	var resp struct {
		Search struct {
			SearchFreelancerPublicProfile FreelancerSearchResult `json:"searchFreelancerPublicProfile"`
		} `json:"search"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.Search.SearchFreelancerPublicProfile, nil
}

// UpdateAvailabilityInput represents input for updating availability
type UpdateAvailabilityInput struct {
	Availability string `json:"availability"`
}

// UpdateFreelancerAvailability updates freelancer availability
func (s *FreelancersService) UpdateFreelancerAvailability(ctx context.Context, input UpdateAvailabilityInput) error {
	mutation := `
		mutation UpdateFreelancerAvailability($input: UpdateFreelancerAvailabilityInput!) {
			updateFreelancerAvailability(input: $input) {
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
		UpdateFreelancerAvailability struct {
			Success bool `json:"success"`
		} `json:"updateFreelancerAvailability"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	return nil
}