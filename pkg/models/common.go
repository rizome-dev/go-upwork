// Package models contains shared data models for the Upwork SDK.
package models

import "time"

// ID represents a GraphQL ID type
type ID string

// Money represents a monetary value
type Money struct {
	RawValue     float64 `json:"rawValue"`
	Currency     string  `json:"currency"`
	DisplayValue string  `json:"displayValue"`
}

// DateTime represents a date/time value
type DateTime struct {
	RawValue     string `json:"rawValue"`
	DisplayValue string `json:"displayValue"`
}

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

// PaginationInput represents pagination parameters
type PaginationInput struct {
	First  int    `json:"first,omitempty"`
	After  string `json:"after,omitempty"`
	Last   int    `json:"last,omitempty"`
	Before string `json:"before,omitempty"`
}

// SortOrder represents sort order
type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

// DateRange represents a date range
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Location represents a location
type Location struct {
	Country     string `json:"country"`
	State       string `json:"state"`
	City        string `json:"city"`
	Timezone    string `json:"timezone"`
	OffsetToUTC int    `json:"offsetToUTC"`
}