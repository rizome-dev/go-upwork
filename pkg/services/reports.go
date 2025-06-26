package services

import (
	"github.com/rizome-dev/go-upwork/pkg/models"
	"context"
	"time"
)

// ReportsService handles report-related API operations
type ReportsService struct {
	client *BaseClient
}

// NewReportsService creates a new reports service
func NewReportsService(client *BaseClient) *ReportsService {
	return &ReportsService{client: client}
}

// TransactionHistory represents transaction history
type TransactionHistory struct {
	TransactionDetail TransactionDetail `json:"transactionDetail"`
}

// TransactionDetail contains transaction details
type TransactionDetail struct {
	TransactionHistoryRows []TransactionHistoryRow `json:"transactionHistoryRow"`
}

// TransactionHistoryRow represents a single transaction row
type TransactionHistoryRow struct {
	RowNumber                   int       `json:"rowNumber"`
	RecordID                    string    `json:"recordId"`
	Type                        string    `json:"type"`
	AccountingSubtype           string    `json:"accountingSubtype"`
	Description                 string    `json:"description"`
	DescriptionUI               string    `json:"descriptionUI"`
	TransactionCreationDate     DateTime  `json:"transactionCreationDate"`
	TransactionReviewDueDate    DateTime  `json:"transactionReviewDueDate"`
	TransactionAmount           Money     `json:"transactionAmount"`
	AmountCreditedToUser        Money     `json:"amountCreditedToUser"`
	Payment                     Money     `json:"payment"`
	PaymentStatus               string    `json:"paymentStatus"`
	RelatedAssignment           string    `json:"relatedAssignment"`
	RelatedAccountingEntity     string    `json:"relatedAccountingEntity"`
	RelatedTransactionID        string    `json:"relatedTransactionId"`
	RelatedInvoiceID            string    `json:"relatedInvoiceId"`
	PurchaseOrderNumber         string    `json:"purchaseOrderNumber"`
	AssignmentTeamCompanyID     string    `json:"assignmentTeamCompanyId"`
	AssignmentTeamCompanyReference string `json:"assignmentTeamCompanyReference"`
	AssignmentCompanyName       string    `json:"assignmentCompanyName"`
	AssignmentDeveloperName     string    `json:"assignmentDeveloperName"`
	AssignmentTeamUserID        string    `json:"assignmentTeamUserId"`
	AssignmentTeamUserReference string    `json:"assignmentTeamUserReference"`
}

// TransactionHistoryInput represents input for transaction history query
type TransactionHistoryInput struct {
	AccountingEntityIDs []string    `json:"aceIds"`
	DateRange           DateRange   `json:"transactionDateTime"`
}

// DateRange represents a date range
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// GetTransactionHistory retrieves transaction history
func (s *ReportsService) GetTransactionHistory(ctx context.Context, input TransactionHistoryInput) (*TransactionHistory, error) {
	query := `
		query TransactionHistory($aceIds_any: [ID!]!, $transactionDateTime_bt: DateTimeRange!) {
			transactionHistory(
				transactionHistoryFilter: {
					aceIds_any: $aceIds_any,
					transactionDateTime_bt: $transactionDateTime_bt
				}
			) {
				transactionDetail {
					transactionHistoryRow {
						rowNumber
						recordId
						type
						accountingSubtype
						description
						descriptionUI
						transactionCreationDate
						transactionReviewDueDate
						transactionAmount {
							rawValue
							currency
							displayValue
						}
						amountCreditedToUser {
							rawValue
							currency
							displayValue
						}
						payment {
							rawValue
							currency
							displayValue
						}
						paymentStatus
						relatedAssignment
						relatedAccountingEntity
						relatedTransactionId
						relatedInvoiceId
						purchaseOrderNumber
						assignmentTeamCompanyId
						assignmentTeamCompanyReference
						assignmentCompanyName
						assignmentDeveloperName
						assignmentTeamUserId
						assignmentTeamUserReference
					}
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"aceIds_any":            input.AccountingEntityIDs,
			"transactionDateTime_bt": input.DateRange,
		},
	}
	
	var resp struct {
		TransactionHistory TransactionHistory `json:"transactionHistory"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.TransactionHistory, nil
}

// TimeReport represents a time report
type TimeReport struct {
	DateWorkedOn        DateTime `json:"dateWorkedOn"`
	WeekWorkedOn        DateTime `json:"weekWorkedOn"`
	MonthWorkedOn       int      `json:"monthWorkedOn"`
	YearWorkedOn        int      `json:"yearWorkedOn"`
	Freelancer          User     `json:"freelancer"`
	Team                Team     `json:"team"`
	Contract            Contract `json:"contract"`
	Task                string   `json:"task"`
	TaskDescription     string   `json:"taskDescription"`
	Memo                string   `json:"memo"`
	TotalHoursWorked    float64  `json:"totalHoursWorked"`
	TotalCharges        Money    `json:"totalCharges"`
	TotalOnlineHoursWorked  float64  `json:"totalOnlineHoursWorked"`
	TotalOnlineCharge   Money    `json:"totalOnlineCharge"`
	TotalOfflineHoursWorked float64  `json:"totalOfflineHoursWorked"`
	TotalOfflineCharge  Money    `json:"totalOfflineCharge"`
}

// TimeReportList represents a list of time reports
type TimeReportList struct {
	TotalCount int              `json:"totalCount"`
	PageInfo   PageInfo         `json:"pageInfo"`
	Edges      []TimeReportEdge `json:"edges"`
}

// TimeReportEdge represents a time report edge
type TimeReportEdge struct {
	Cursor string     `json:"cursor"`
	Node   TimeReport `json:"node"`
}

// TimeReportInput represents input for time report query
type TimeReportInput struct {
	OrganizationID string           `json:"organizationId"`
	DateRange      DateRange        `json:"timeReportDate"`
	Pagination     *PaginationInput `json:"pagination,omitempty"`
}

// GetTimeReport retrieves time reports
func (s *ReportsService) GetTimeReport(ctx context.Context, input TimeReportInput) (*TimeReportList, error) {
	query := `
		query TimeReport($orgId: ID!, $after: String, $first: Int!, $timeReportDate_bt: DateTimeRange!) {
			contractTimeReport(
				filter: {
					organizationId_eq: $orgId,
					timeReportDate_bt: $timeReportDate_bt
				}
				pagination: {after: $after, first: $first}
			) {
				totalCount
				pageInfo {
					hasNextPage
					endCursor
				}
				edges {
					cursor
					node {
						dateWorkedOn
						weekWorkedOn
						monthWorkedOn
						yearWorkedOn
						freelancer {
							id
							nid
							name
						}
						team {
							id
							name
						}
						contract {
							id
						}
						task
						taskDescription
						memo
						totalHoursWorked
						totalCharges
						totalOnlineHoursWorked
						totalOnlineCharge
						totalOfflineHoursWorked
						totalOfflineCharge
					}
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"orgId":              input.OrganizationID,
		"timeReportDate_bt":  input.DateRange,
	}
	
	if input.Pagination != nil {
		variables["after"] = input.Pagination.After
		variables["first"] = input.Pagination.First
	} else {
		variables["first"] = 50 // Default page size
	}
	
	req := &GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	
	var resp struct {
		ContractTimeReport TimeReportList `json:"contractTimeReport"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.ContractTimeReport, nil
}

// WorkDiary represents work diary data
type WorkDiary struct {
	Total     int                `json:"total"`
	Snapshots []WorkDiarySnapshot `json:"snapshots"`
}

// WorkDiarySnapshot represents a work diary snapshot
type WorkDiarySnapshot struct {
	Contract            WorkDiaryContract    `json:"contract"`
	User                User                 `json:"user"`
	Duration            string               `json:"duration"`
	DurationInt         int                  `json:"durationInt"`
	Task                Task                 `json:"task"`
	Time                WorkDiaryTime        `json:"time"`
	Screenshots         []Screenshot         `json:"screenshots"`
}

// WorkDiaryContract represents contract info in work diary
type WorkDiaryContract struct {
	ID            string `json:"id"`
	ContractTitle string `json:"contractTitle"`
	UserID        string `json:"userId"`
}

// Task represents a task
type Task struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Memo        string `json:"memo"`
}

// WorkDiaryTime represents time info in work diary
type WorkDiaryTime struct {
	TrackedTime    string `json:"trackedTime"`
	ManualTime     string `json:"manualTime"`
	Overtime       string `json:"overtime"`
	FirstWorked    string `json:"firstWorked"`
	LastWorked     string `json:"lastWorked"`
	FirstWorkedInt int    `json:"firstWorkedInt"`
	LastWorkedInt  int    `json:"lastWorkedInt"`
	LastScreenshot string `json:"lastScreenshot"`
}

// Screenshot represents a screenshot
type Screenshot struct {
	Activity               int    `json:"activity"`
	ScreenshotURL          string `json:"screenshotUrl"`
	ScreenshotImage        string `json:"screenshotImage"`
	ScreenshotImageLarge   string `json:"screenshotImageLarge"`
	ScreenshotImageMedium  string `json:"screenshotImageMedium"`
	ScreenshotImageThumbnail string `json:"screenshotImageThumbnail"`
	HasWebcam              bool   `json:"hasWebcam"`
	HasScreenshot          bool   `json:"hasScreenshot"`
	WebcamURL              string `json:"webcamUrl"`
	WebcamImage            string `json:"webcamImage"`
	WebcamImageThumbnail   string `json:"webcamImageThumbnail"`
}

// GetWorkDiaryByCompany retrieves work diary for a company
func (s *ReportsService) GetWorkDiaryByCompany(ctx context.Context, companyID string, date string) (*WorkDiary, error) {
	query := `
		query GetWorkDiaryCompany($companyId: ID!, $date: String!) {
			workDiaryCompany(workDiaryCompanyInput: {companyId: $companyId, date: $date}) {
				total
				snapshots {
					contract {
						id
						contractTitle
						userId
					}
					user {
						id
						name
						portraitUrl
					}
					duration
					durationInt
					task {
						id
						code
						description
						memo
					}
					time {
						trackedTime
						manualTime
						overtime
						firstWorked
						lastWorked
						firstWorkedInt
						lastWorkedInt
						lastScreenshot
					}
					screenshots {
						activity
						screenshotUrl
						screenshotImage
						screenshotImageLarge
						screenshotImageMedium
						screenshotImageThumbnail
						hasWebcam
						hasScreenshot
						webcamUrl
						webcamImage
						webcamImageThumbnail
					}
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"companyId": companyID,
			"date":      date,
		},
	}
	
	var resp struct {
		WorkDiaryCompany WorkDiary `json:"workDiaryCompany"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.WorkDiaryCompany, nil
}