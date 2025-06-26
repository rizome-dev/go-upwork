package services

import (
	"github.com/rizome-dev/go-upwork/pkg/models"
	"context"
	"fmt"
)

// JobsService handles job-related API operations
type JobsService struct {
	client *BaseClient
}

// NewJobsService creates a new jobs service
func NewJobsService(client *BaseClient) *JobsService {
	return &JobsService{client: client}
}

// JobPosting represents a job posting
type JobPosting struct {
	ID               ID                    `json:"id"`
	Content          JobContent            `json:"content"`
	Info             JobInfo               `json:"info"`
	ContractTerms    ContractTerms         `json:"contractTerms"`
	Classification   JobClassification     `json:"classification"`
	Ownership        JobOwnership          `json:"ownership"`
	Visibility       string                `json:"visibility"`
	Attachment       *Attachment           `json:"attachment"`
	ContractorSelection ContractorSelection `json:"contractorSelection"`
}

// JobInfo represents job information
type JobInfo struct {
	Status           JobStatus  `json:"status"`
	HourlyBudgetMin  *Money     `json:"hourlyBudgetMin"`
	HourlyBudgetMax  *Money     `json:"hourlyBudgetMax"`
	AuditTime        AuditTime  `json:"auditTime"`
	FilledDateTime   *DateTime  `json:"filledDateTime"`
	LegacyCiphertext string     `json:"legacyCiphertext"`
	KeepOpenOnHire   bool       `json:"keepOpenOnHire"`
	SiteSource       string     `json:"siteSource"`
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusOpen      JobStatus = "OPEN"
	JobStatusFilled    JobStatus = "FILLED"
	JobStatusCancelled JobStatus = "CANCELLED"
	JobStatusDraft     JobStatus = "DRAFT"
)

// AuditTime represents audit timestamps
type AuditTime struct {
	CreatedDateTime  DateTime `json:"createdDateTime"`
	ModifiedDateTime DateTime `json:"modifiedDateTime"`
}

// ContractTerms represents contract terms for a job
type ContractTerms struct {
	ContractType           ContractType            `json:"contractType"`
	ContractStartDate      *DateTime               `json:"contractStartDate"`
	ContractEndDate        *DateTime               `json:"contractEndDate"`
	HourlyContractTerms    *HourlyContractTerms    `json:"hourlyContractTerms"`
	FixedPriceContractTerms *FixedPriceContractTerms `json:"fixedPriceContractTerms"`
}

// HourlyContractTerms represents hourly contract terms
type HourlyContractTerms struct {
	EngagementDuration EngagementDuration `json:"engagementDuration"`
	EngagementType     EngagementType     `json:"engagementType"`
}

// FixedPriceContractTerms represents fixed price contract terms
type FixedPriceContractTerms struct {
	EngagementDuration EngagementDuration `json:"engagementDuration"`
}

// EngagementDuration represents engagement duration
type EngagementDuration struct {
	ID    ID     `json:"id"`
	Weeks int    `json:"weeks"`
	Label string `json:"label"`
}

// EngagementType represents engagement type
type EngagementType string

const (
	EngagementTypeAsNeeded EngagementType = "AS_NEEDED"
	EngagementTypePartTime EngagementType = "PART_TIME"
	EngagementTypeFullTime EngagementType = "FULL_TIME"
)

// JobClassification represents job classification
type JobClassification struct {
	Category    Category   `json:"category"`
	SubCategory SubCategory `json:"subCategory"`
	Skills      []Skill    `json:"skills"`
}

// Category represents a job category
type Category struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
}

// SubCategory represents a job subcategory
type SubCategory struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
}

// Skill represents a skill
type Skill struct {
	ID         ID     `json:"id"`
	PrettyName string `json:"prettyName"`
}

// JobOwnership represents job ownership information
type JobOwnership struct {
	Company Company `json:"company"`
	Team    Team    `json:"team"`
}

// Team represents a team
type Team struct {
	ID   ID     `json:"id"`
	Rid  string `json:"rid"`
	Name string `json:"name"`
}

// Attachment represents an attachment
type Attachment struct {
	Link string `json:"link"`
}

// ContractorSelection represents contractor selection criteria
type ContractorSelection struct {
	Qualification ContractorQualification `json:"qualification"`
}

// ContractorQualification represents contractor qualification
type ContractorQualification struct {
	ContractorType string `json:"contractorType"`
}

// CreateJobPostingInput represents input for creating a job
type CreateJobPostingInput struct {
	Title               string              `json:"title"`
	Description         string              `json:"description"`
	CategoryID          string              `json:"categoryId"`
	SubCategoryID       string              `json:"subCategoryId"`
	Skills              []string            `json:"skills"`
	ContractType        ContractType        `json:"contractType"`
	HourlyBudgetMin     *float64            `json:"hourlyBudgetMin,omitempty"`
	HourlyBudgetMax     *float64            `json:"hourlyBudgetMax,omitempty"`
	FixedPriceBudget    *float64            `json:"fixedPriceBudget,omitempty"`
	Duration            string              `json:"duration,omitempty"`
	Workload            string              `json:"workload,omitempty"`
	ContractorType      string              `json:"contractorType,omitempty"`
	TeamID              string              `json:"teamId"`
}

// CreateJobPosting creates a new job posting
func (s *JobsService) CreateJobPosting(ctx context.Context, input CreateJobPostingInput) (*JobPosting, error) {
	mutation := `
		mutation CreateJobPosting($input: CreateJobPostingInput!) {
			createJobPosting(input: $input) {
				id
				content {
					title
					description
				}
				info {
					status
					legacyCiphertext
					auditTime {
						createdDateTime
					}
				}
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
		CreateJobPosting JobPosting `json:"createJobPosting"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.CreateJobPosting, nil
}

// UpdateJobPostingInput represents input for updating a job
type UpdateJobPostingInput struct {
	ID               string   `json:"id"`
	Title            string   `json:"title,omitempty"`
	Description      string   `json:"description,omitempty"`
	Skills           []string `json:"skills,omitempty"`
	HourlyBudgetMin  *float64 `json:"hourlyBudgetMin,omitempty"`
	HourlyBudgetMax  *float64 `json:"hourlyBudgetMax,omitempty"`
	FixedPriceBudget *float64 `json:"fixedPriceBudget,omitempty"`
}

// UpdateJobPosting updates an existing job posting
func (s *JobsService) UpdateJobPosting(ctx context.Context, input UpdateJobPostingInput) (*JobPosting, error) {
	mutation := `
		mutation UpdateJobPosting($input: UpdateJobPostingInput!) {
			updateJobPosting(input: $input) {
				id
				content {
					title
					description
				}
				info {
					status
					auditTime {
						modifiedDateTime
					}
				}
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
		UpdateJobPosting JobPosting `json:"updateJobPosting"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.UpdateJobPosting, nil
}

// GetJobPosting retrieves a job posting by ID
func (s *JobsService) GetJobPosting(ctx context.Context, jobID string) (*JobPosting, error) {
	query := `
		query GetJobPosting($jobPostingId: ID!) {
			jobPosting(jobPostingId: $jobPostingId) {
				id
				content {
					title
					description
				}
				info {
					status
					hourlyBudgetMin {
						rawValue
						currency
					}
					hourlyBudgetMax {
						rawValue
						currency
					}
					auditTime {
						createdDateTime
						modifiedDateTime
					}
					filledDateTime
					legacyCiphertext
					keepOpenOnHire
				}
				contractTerms {
					contractType
					contractStartDate
					contractEndDate
					hourlyContractTerms {
						engagementDuration {
							id
							weeks
							label
						}
						engagementType
					}
					fixedPriceContractTerms {
						engagementDuration {
							id
							weeks
							label
						}
					}
				}
				classification {
					category {
						id
						name
					}
					subCategory {
						id
						name
					}
					skills {
						id
						prettyName
					}
				}
				ownership {
					company {
						id
						name
					}
					team {
						id
						rid
						name
					}
				}
				visibility
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"jobPostingId": jobID,
		},
	}
	
	var resp struct {
		JobPosting JobPosting `json:"jobPosting"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.JobPosting, nil
}

// ListJobsInput represents input for listing jobs
type ListJobsInput struct {
	TeamIDs      []string         `json:"postByTeamIds,omitempty"`
	PersonIDs    []string         `json:"postByPersonIds,omitempty"`
	Status       []JobStatus      `json:"status,omitempty"`
	ContractType ContractType     `json:"contractType,omitempty"`
	CreatedFrom  string           `json:"createdDateTimeFrom,omitempty"`
	CreatedTo    string           `json:"createdDateTimeTo,omitempty"`
	Pagination   *PaginationInput `json:"pagination,omitempty"`
}

// JobPostingList represents a list of job postings
type JobPostingList struct {
	TotalCount int              `json:"totalCount"`
	PageInfo   PageInfo         `json:"pageInfo"`
	Edges      []JobPostingEdge `json:"edges"`
}

// JobPostingEdge represents a job posting edge
type JobPostingEdge struct {
	Cursor string     `json:"cursor"`
	Node   JobPosting `json:"node"`
}

// ListJobs returns a list of jobs
func (s *JobsService) ListJobs(ctx context.Context, input ListJobsInput) (*JobPostingList, error) {
	query := `
		query ListJobs($filter: JobPostingFilterInput, $sortAttribute: JobPostingSortAttribute) {
			organization {
				jobPosting(jobPostingFilter: $filter, sortAttribute: $sortAttribute) {
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
							content {
								title
							}
							info {
								status
								auditTime {
									createdDateTime
								}
							}
							contractTerms {
								contractType
							}
						}
					}
				}
			}
		}
	`
	
	filter := map[string]interface{}{}
	if len(input.TeamIDs) > 0 {
		filter["postByTeamIds_any"] = input.TeamIDs
	}
	if len(input.PersonIDs) > 0 {
		filter["postByPersonIds_any"] = input.PersonIDs
	}
	if input.CreatedFrom != "" {
		filter["createdDateTimeFrom_eq"] = input.CreatedFrom
	}
	if input.CreatedTo != "" {
		filter["createdDateTimeTo_eq"] = input.CreatedTo
	}
	if input.Pagination != nil {
		filter["pagination_eq"] = input.Pagination
	}
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"filter": filter,
		},
	}
	
	var resp struct {
		Organization struct {
			JobPosting JobPostingList `json:"jobPosting"`
		} `json:"organization"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.Organization.JobPosting, nil
}

// MarketplaceJobFilter represents marketplace job search filters
type MarketplaceJobFilter struct {
	SearchExpression string           `json:"searchExpression_eq,omitempty"`
	SkillExpression  string           `json:"skillExpression_eq,omitempty"`
	TitleExpression  string           `json:"titleExpression_eq,omitempty"`
	CategoryIDs      []string         `json:"categoryIds_any,omitempty"`
	SubcategoryIDs   []string         `json:"subcategoryIds_any,omitempty"`
	JobType          ContractType     `json:"jobType_eq,omitempty"`
	Duration         string           `json:"duration_eq,omitempty"`
	Workload         string           `json:"workload_eq,omitempty"`
	ExperienceLevel  string           `json:"experienceLevel_eq,omitempty"`
	DaysPosted       int              `json:"daysPosted_eq,omitempty"`
	Pagination       *PaginationInput `json:"pagination_eq,omitempty"`
}

// SearchJobs searches for jobs in the marketplace
func (s *JobsService) SearchJobs(ctx context.Context, filter MarketplaceJobFilter) (*JobPostingList, error) {
	query := `
		query SearchJobs($filter: MarketplaceJobFilter) {
			marketplaceJobPostings(marketPlaceJobFilter: $filter) {
				totalCount
				pageInfo {
					hasNextPage
					endCursor
				}
				edges {
					cursor
					node {
						id
						title
						description
						createdDateTime
						client {
							location {
								country
							}
							totalFeedback
							totalHires
							totalPostedJobs
						}
					}
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"filter": filter,
		},
	}
	
	var resp struct {
		MarketplaceJobPostings JobPostingList `json:"marketplaceJobPostings"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.MarketplaceJobPostings, nil
}