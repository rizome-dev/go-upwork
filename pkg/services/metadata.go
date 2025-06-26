package services

import (
	"github.com/rizome-dev/go-upwork/pkg/models"
	"context"
)

// MetadataService handles metadata-related API operations
type MetadataService struct {
	client *BaseClient
}

// NewMetadataService creates a new metadata service
func NewMetadataService(client *BaseClient) *MetadataService {
	return &MetadataService{client: client}
}

// OntologyCategory represents an ontology category
type OntologyCategory struct {
	ID             ID                   `json:"id"`
	PreferredLabel string               `json:"preferredLabel"`
	AltLabel       string               `json:"altLabel"`
	Slug           string               `json:"slug"`
	OntologyID     string               `json:"ontologyId"`
	Subcategories  []OntologySubcategory `json:"subcategories"`
	Services       []OntologyService    `json:"services"`
}

// OntologySubcategory represents an ontology subcategory
type OntologySubcategory struct {
	ID             ID     `json:"id"`
	PreferredLabel string `json:"preferredLabel"`
	AltLabel       string `json:"altLabel"`
	Slug           string `json:"slug"`
}

// OntologyService represents an ontology service
type OntologyService struct {
	ID             ID     `json:"id"`
	PreferredLabel string `json:"preferredLabel"`
}

// OntologySkill represents an ontology skill
type OntologySkill struct {
	ID             ID     `json:"id"`
	PreferredLabel string `json:"preferredLabel"`
}

// Region represents a geographical region
type Region struct {
	ID           ID      `json:"id"`
	Name         string  `json:"name"`
	ParentRegion *Region `json:"parentRegion"`
}

// Country represents a country
type Country struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// Language represents a language
type Language struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// Reason represents a reason (for various actions)
type Reason struct {
	ID     ID     `json:"id"`
	Reason string `json:"reason"`
	Alias  string `json:"alias"`
}

// ReasonType represents the type of reason
type ReasonType string

const (
	ReasonTypeJobPostingClose ReasonType = "JOB_POSTING_CLOSE"
	ReasonTypeContractEnd     ReasonType = "CONTRACT_END"
	ReasonTypeProposalDecline ReasonType = "PROPOSAL_DECLINE"
)

// GetCategories returns all ontology categories
func (s *MetadataService) GetCategories(ctx context.Context) ([]OntologyCategory, error) {
	query := `
		query GetOntologyCategories {
			ontologyCategories {
				id
				preferredLabel
				altLabel
				slug
				ontologyId
				subcategories {
					id
					preferredLabel
					altLabel
					slug
				}
				services {
					id
					preferredLabel
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
	}
	
	var resp struct {
		OntologyCategories []OntologyCategory `json:"ontologyCategories"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return resp.OntologyCategories, nil
}

// GetSkills returns ontology skills with pagination
func (s *MetadataService) GetSkills(ctx context.Context, limit int, offset int) ([]OntologySkill, error) {
	query := `
		query GetOntologySkills($limit: Int!, $offset: Int) {
			ontologyBrowserSkills(limit: $limit, offset: $offset) {
				id
				preferredLabel
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
	}
	
	var resp struct {
		OntologyBrowserSkills []OntologySkill `json:"ontologyBrowserSkills"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return resp.OntologyBrowserSkills, nil
}

// GetRegions returns all regions
func (s *MetadataService) GetRegions(ctx context.Context) ([]Region, error) {
	query := `
		query GetRegions {
			regions {
				id
				name
				parentRegion {
					id
					name
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
	}
	
	var resp struct {
		Regions []Region `json:"regions"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return resp.Regions, nil
}

// GetCountries returns all countries
func (s *MetadataService) GetCountries(ctx context.Context) ([]Country, error) {
	query := `
		query GetCountries {
			countries {
				id
				name
				code
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
	}
	
	var resp struct {
		Countries []Country `json:"countries"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return resp.Countries, nil
}

// GetLanguages returns all languages
func (s *MetadataService) GetLanguages(ctx context.Context) ([]Language, error) {
	query := `
		query GetLanguages {
			languages {
				id
				name
				code
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
	}
	
	var resp struct {
		Languages []Language `json:"languages"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return resp.Languages, nil
}

// GetReasons returns reasons by type
func (s *MetadataService) GetReasons(ctx context.Context, reasonType ReasonType, all bool) ([]Reason, error) {
	query := `
		query GetReasons($reasonType: ReasonType!, $all: Boolean) {
			reasons(reasonType: $reasonType, all: $all) {
				id
				reason
				alias
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"reasonType": reasonType,
			"all":        all,
		},
	}
	
	var resp struct {
		Reasons []Reason `json:"reasons"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return resp.Reasons, nil
}

// TimeZone represents a time zone
type TimeZone struct {
	ID     ID     `json:"id"`
	Name   string `json:"name"`
	Offset int    `json:"offset"`
}

// GetTimeZones returns all time zones
func (s *MetadataService) GetTimeZones(ctx context.Context) ([]TimeZone, error) {
	query := `
		query GetTimeZones {
			timeZones {
				id
				name
				offset
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
	}
	
	var resp struct {
		TimeZones []TimeZone `json:"timeZones"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return resp.TimeZones, nil
}

// SearchSkillsInput represents input for searching skills
type SearchSkillsInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// SearchSkills searches for skills by query
func (s *MetadataService) SearchSkills(ctx context.Context, input SearchSkillsInput) ([]OntologySkill, error) {
	query := `
		query SearchSkills($query: String!, $limit: Int!) {
			ontologyElementsSearchByPrefLabel(
				prefLabel: $query,
				elementType: "skill",
				limit: $limit
			) {
				id
				preferredLabel
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"query": input.Query,
			"limit": input.Limit,
		},
	}
	
	var resp struct {
		OntologyElementsSearchByPrefLabel []OntologySkill `json:"ontologyElementsSearchByPrefLabel"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return resp.OntologyElementsSearchByPrefLabel, nil
}