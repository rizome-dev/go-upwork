package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/rizome-dev/go-upwork/internal/graphql"
	"github.com/rizome-dev/go-upwork/pkg/models"
	"github.com/rizome-dev/go-upwork/tests/mocks"
	"github.com/rizome-dev/go-upwork/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupContractsService(responses ...mocks.MockResponse) (*ContractsService, *mocks.RequestRecorder) {
	recorder := mocks.NewRequestRecorder(responses...)
	client, _ := graphql.NewClient("https://api.upwork.com/graphql", recorder)
	rateLimiter := mocks.NewMockRateLimiter()
	service := NewContractsService(client, rateLimiter)
	return service, recorder
}

func TestGetContract(t *testing.T) {
	tests := []struct {
		name         string
		contractID   string
		mockResponse interface{}
		wantErr      bool
		validate     func(t *testing.T, contract *models.Contract)
	}{
		{
			name:       "successful get contract",
			contractID: "123456789",
			mockResponse: testutils.MockGraphQLResponse(
				map[string]interface{}{
					"contractByTerm": map[string]interface{}{
						"id":     "123456789",
						"title":  "Test Contract",
						"status": "ACTIVE",
						"startDate": "2024-01-01T00:00:00Z",
						"hourlyRate": map[string]interface{}{
							"amount":   "50.00",
							"currency": "USD",
						},
						"client": map[string]interface{}{
							"id":   "client_123",
							"name": "Test Client",
							"company": map[string]interface{}{
								"id":   "company_123",
								"name": "Test Company",
							},
						},
						"freelancer": map[string]interface{}{
							"id":    "freelancer_456",
							"name":  "Test Freelancer",
							"email": "freelancer@example.com",
						},
						"milestones": []interface{}{
							map[string]interface{}{
								"id":     "milestone_1",
								"title":  "First Milestone",
								"amount": map[string]interface{}{
									"amount":   "500.00",
									"currency": "USD",
								},
								"status": "ACTIVE",
							},
						},
					},
				},
				nil,
			),
			wantErr: false,
			validate: func(t *testing.T, contract *models.Contract) {
				assert.Equal(t, "123456789", contract.ID)
				assert.Equal(t, "Test Contract", contract.Title)
				assert.Equal(t, models.ContractStatusActive, contract.Status)
				assert.NotNil(t, contract.HourlyRate)
				assert.Equal(t, 50.00, contract.HourlyRate.Amount)
				assert.Equal(t, "USD", contract.HourlyRate.Currency)
				assert.NotNil(t, contract.Client)
				assert.Equal(t, "client_123", contract.Client.ID)
				assert.Equal(t, "Test Client", contract.Client.Name)
				assert.NotNil(t, contract.Freelancer)
				assert.Equal(t, "freelancer_456", contract.Freelancer.ID)
				assert.Len(t, contract.Milestones, 1)
				assert.Equal(t, "milestone_1", contract.Milestones[0].ID)
			},
		},
		{
			name:       "contract not found",
			contractID: "invalid",
			mockResponse: testutils.MockGraphQLResponse(
				nil,
				[]interface{}{
					testutils.CreateGraphQLError("Contract not found", "NOT_FOUND"),
				},
			),
			wantErr: true,
		},
		{
			name:       "empty contract ID",
			contractID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.mockResponse)
			service, recorder := setupContractsService(
				mocks.MockResponse{
					StatusCode: 200,
					Body:       string(body),
				},
			)

			contract, err := service.GetContract(context.Background(), tt.contractID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, contract)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, contract)
				if tt.validate != nil {
					tt.validate(t, contract)
				}
			}

			// Verify the GraphQL query was correct
			if tt.contractID != "" && len(recorder.Requests) > 0 {
				req := recorder.GetLastRequest()
				body := recorder.GetRequestBody(0)
				
				var gqlReq map[string]interface{}
				err := json.Unmarshal([]byte(body), &gqlReq)
				require.NoError(t, err)
				
				assert.Contains(t, gqlReq["query"], "contractByTerm")
				assert.Equal(t, map[string]interface{}{"id": tt.contractID}, gqlReq["variables"])
			}
		})
	}
}

func TestListContracts(t *testing.T) {
	tests := []struct {
		name         string
		options      models.ContractListOptions
		mockResponse interface{}
		wantErr      bool
		validate     func(t *testing.T, contracts []*models.Contract, pageInfo *models.PageInfo)
	}{
		{
			name: "list active contracts",
			options: models.ContractListOptions{
				Status: models.ContractStatusActive,
				Limit:  10,
			},
			mockResponse: testutils.MockGraphQLResponse(
				map[string]interface{}{
					"vendorContracts": map[string]interface{}{
						"edges": []interface{}{
							map[string]interface{}{
								"node": map[string]interface{}{
									"id":     "c1",
									"title":  "Contract 1",
									"status": "ACTIVE",
								},
							},
							map[string]interface{}{
								"node": map[string]interface{}{
									"id":     "c2",
									"title":  "Contract 2",
									"status": "ACTIVE",
								},
							},
						},
						"pageInfo": map[string]interface{}{
							"hasNextPage": true,
							"endCursor":   "cursor123",
						},
					},
				},
				nil,
			),
			wantErr: false,
			validate: func(t *testing.T, contracts []*models.Contract, pageInfo *models.PageInfo) {
				assert.Len(t, contracts, 2)
				assert.Equal(t, "c1", contracts[0].ID)
				assert.Equal(t, "c2", contracts[1].ID)
				assert.True(t, pageInfo.HasNextPage)
				assert.Equal(t, "cursor123", pageInfo.EndCursor)
			},
		},
		{
			name: "list with pagination",
			options: models.ContractListOptions{
				Limit:  5,
				After:  "prevCursor",
			},
			mockResponse: testutils.MockGraphQLResponse(
				map[string]interface{}{
					"vendorContracts": map[string]interface{}{
						"edges": []interface{}{},
						"pageInfo": map[string]interface{}{
							"hasNextPage": false,
							"endCursor":   "",
						},
					},
				},
				nil,
			),
			wantErr: false,
			validate: func(t *testing.T, contracts []*models.Contract, pageInfo *models.PageInfo) {
				assert.Len(t, contracts, 0)
				assert.False(t, pageInfo.HasNextPage)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.mockResponse)
			service, recorder := setupContractsService(
				mocks.MockResponse{
					StatusCode: 200,
					Body:       string(body),
				},
			)

			contracts, pageInfo, err := service.ListContracts(context.Background(), tt.options)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, contracts, pageInfo)
				}
			}

			// Verify query variables
			if len(recorder.Requests) > 0 {
				body := recorder.GetRequestBody(0)
				var gqlReq map[string]interface{}
				err := json.Unmarshal([]byte(body), &gqlReq)
				require.NoError(t, err)
				
				variables := gqlReq["variables"].(map[string]interface{})
				if tt.options.Limit > 0 {
					assert.Equal(t, float64(tt.options.Limit), variables["first"])
				}
				if tt.options.After != "" {
					assert.Equal(t, tt.options.After, variables["after"])
				}
			}
		})
	}
}

func TestCreateContract(t *testing.T) {
	tests := []struct {
		name         string
		input        models.CreateContractInput
		mockResponse interface{}
		wantErr      bool
		validate     func(t *testing.T, contract *models.Contract)
	}{
		{
			name: "create hourly contract",
			input: models.CreateContractInput{
				Title:         "New Contract",
				FreelancerID:  "freelancer_123",
				HourlyRate:    &models.Money{Amount: 75.00, Currency: "USD"},
				WeeklyLimit:   40,
			},
			mockResponse: testutils.MockGraphQLResponse(
				map[string]interface{}{
					"createContract": map[string]interface{}{
						"contract": map[string]interface{}{
							"id":     "new_contract_123",
							"title":  "New Contract",
							"status": "ACTIVE",
							"hourlyRate": map[string]interface{}{
								"amount":   "75.00",
								"currency": "USD",
							},
						},
						"success": true,
					},
				},
				nil,
			),
			wantErr: false,
			validate: func(t *testing.T, contract *models.Contract) {
				assert.Equal(t, "new_contract_123", contract.ID)
				assert.Equal(t, "New Contract", contract.Title)
				assert.Equal(t, 75.00, contract.HourlyRate.Amount)
			},
		},
		{
			name: "create fixed price contract",
			input: models.CreateContractInput{
				Title:        "Fixed Price Contract",
				FreelancerID: "freelancer_456",
				Milestones: []models.MilestoneInput{
					{
						Title:       "First Milestone",
						Description: "Initial work",
						Amount:      &models.Money{Amount: 1000.00, Currency: "USD"},
						DueDate:     time.Now().Add(7 * 24 * time.Hour),
					},
				},
			},
			mockResponse: testutils.MockGraphQLResponse(
				map[string]interface{}{
					"createContract": map[string]interface{}{
						"contract": map[string]interface{}{
							"id":     "fixed_contract_456",
							"title":  "Fixed Price Contract",
							"status": "ACTIVE",
							"milestones": []interface{}{
								map[string]interface{}{
									"id":     "milestone_new",
									"title":  "First Milestone",
									"amount": map[string]interface{}{
										"amount":   "1000.00",
										"currency": "USD",
									},
								},
							},
						},
						"success": true,
					},
				},
				nil,
			),
			wantErr: false,
			validate: func(t *testing.T, contract *models.Contract) {
				assert.Equal(t, "fixed_contract_456", contract.ID)
				assert.Len(t, contract.Milestones, 1)
				assert.Equal(t, 1000.00, contract.Milestones[0].Amount.Amount)
			},
		},
		{
			name: "validation error",
			input: models.CreateContractInput{
				Title: "", // Invalid empty title
			},
			mockResponse: testutils.MockGraphQLResponse(
				nil,
				[]interface{}{
					testutils.CreateGraphQLError("Title is required", "VALIDATION_ERROR"),
				},
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.mockResponse)
			service, _ := setupContractsService(
				mocks.MockResponse{
					StatusCode: 200,
					Body:       string(body),
				},
			)

			contract, err := service.CreateContract(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, contract)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, contract)
				if tt.validate != nil {
					tt.validate(t, contract)
				}
			}
		})
	}
}

func TestUpdateContract(t *testing.T) {
	tests := []struct {
		name         string
		contractID   string
		input        models.UpdateContractInput
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:       "update hourly limit",
			contractID: "contract_123",
			input: models.UpdateContractInput{
				WeeklyLimit: ptr(30),
			},
			mockResponse: testutils.MockGraphQLResponse(
				map[string]interface{}{
					"updateContractHourlyLimit": map[string]interface{}{
						"contract": map[string]interface{}{
							"id":          "contract_123",
							"weeklyLimit": 30,
						},
						"success": true,
					},
				},
				nil,
			),
			wantErr: false,
		},
		{
			name:       "update contract title",
			contractID: "contract_456",
			input: models.UpdateContractInput{
				Title: ptr("Updated Title"),
			},
			mockResponse: testutils.MockGraphQLResponse(
				map[string]interface{}{
					"updateContract": map[string]interface{}{
						"contract": map[string]interface{}{
							"id":    "contract_456",
							"title": "Updated Title",
						},
						"success": true,
					},
				},
				nil,
			),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.mockResponse)
			service, _ := setupContractsService(
				mocks.MockResponse{
					StatusCode: 200,
					Body:       string(body),
				},
			)

			contract, err := service.UpdateContract(context.Background(), tt.contractID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, contract)
			}
		})
	}
}

func TestPauseContract(t *testing.T) {
	mockResponse := testutils.MockGraphQLResponse(
		map[string]interface{}{
			"pauseContract": map[string]interface{}{
				"contract": map[string]interface{}{
					"id":     "contract_123",
					"status": "PAUSED",
				},
				"success": true,
			},
		},
		nil,
	)

	body, _ := json.Marshal(mockResponse)
	service, recorder := setupContractsService(
		mocks.MockResponse{
			StatusCode: 200,
			Body:       string(body),
		},
	)

	err := service.PauseContract(context.Background(), "contract_123", "Taking a break")
	assert.NoError(t, err)

	// Verify the mutation was called with correct parameters
	reqBody := recorder.GetRequestBody(0)
	var gqlReq map[string]interface{}
	json.Unmarshal([]byte(reqBody), &gqlReq)
	
	variables := gqlReq["variables"].(map[string]interface{})
	assert.Equal(t, "contract_123", variables["contractId"])
	assert.Equal(t, "Taking a break", variables["reason"])
}

func TestEndContract(t *testing.T) {
	tests := []struct {
		name         string
		contractID   string
		reason       string
		feedback     *models.ContractFeedback
		mockResponse interface{}
		wantErr      bool
	}{
		{
			name:       "end contract with feedback",
			contractID: "contract_123",
			reason:     "Project completed",
			feedback: &models.ContractFeedback{
				Score:   5,
				Comment: "Great work!",
				Skills: []models.SkillFeedback{
					{
						SkillID: "skill_1",
						Score:   5,
					},
				},
			},
			mockResponse: testutils.MockGraphQLResponse(
				map[string]interface{}{
					"endContractByClient": map[string]interface{}{
						"contract": map[string]interface{}{
							"id":     "contract_123",
							"status": "ENDED",
						},
						"success": true,
					},
				},
				nil,
			),
			wantErr: false,
		},
		{
			name:       "end contract without feedback",
			contractID: "contract_456",
			reason:     "Budget constraints",
			mockResponse: testutils.MockGraphQLResponse(
				map[string]interface{}{
					"endContractByClient": map[string]interface{}{
						"contract": map[string]interface{}{
							"id":     "contract_456",
							"status": "ENDED",
						},
						"success": true,
					},
				},
				nil,
			),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.mockResponse)
			service, recorder := setupContractsService(
				mocks.MockResponse{
					StatusCode: 200,
					Body:       string(body),
				},
			)

			err := service.EndContract(context.Background(), tt.contractID, tt.reason, tt.feedback)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify feedback was included if provided
				if tt.feedback != nil {
					reqBody := recorder.GetRequestBody(0)
					var gqlReq map[string]interface{}
					json.Unmarshal([]byte(reqBody), &gqlReq)
					
					variables := gqlReq["variables"].(map[string]interface{})
					feedbackData := variables["feedback"].(map[string]interface{})
					assert.Equal(t, float64(tt.feedback.Score), feedbackData["score"])
					assert.Equal(t, tt.feedback.Comment, feedbackData["comment"])
				}
			}
		})
	}
}

func TestGetContractMilestones(t *testing.T) {
	mockResponse := testutils.MockGraphQLResponse(
		map[string]interface{}{
			"contract": map[string]interface{}{
				"milestones": []interface{}{
					map[string]interface{}{
						"id":     "milestone_1",
						"title":  "First Milestone",
						"status": "ACTIVE",
						"amount": map[string]interface{}{
							"amount":   "500.00",
							"currency": "USD",
						},
					},
					map[string]interface{}{
						"id":     "milestone_2",
						"title":  "Second Milestone",
						"status": "SUBMITTED",
						"amount": map[string]interface{}{
							"amount":   "750.00",
							"currency": "USD",
						},
					},
				},
			},
		},
		nil,
	)

	body, _ := json.Marshal(mockResponse)
	service, _ := setupContractsService(
		mocks.MockResponse{
			StatusCode: 200,
			Body:       string(body),
		},
	)

	milestones, err := service.GetContractMilestones(context.Background(), "contract_123")
	assert.NoError(t, err)
	assert.Len(t, milestones, 2)
	assert.Equal(t, "milestone_1", milestones[0].ID)
	assert.Equal(t, "ACTIVE", milestones[0].Status)
	assert.Equal(t, 500.00, milestones[0].Amount.Amount)
}

// Helper function for creating pointers
func ptr[T any](v T) *T {
	return &v
}