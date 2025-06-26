package services

import (
	"github.com/rizome-dev/go-upwork/pkg/models"
	"context"
	"fmt"
	"time"
)

// Milestone represents a milestone in a contract
type Milestone struct {
	ID                     ID               `json:"id"`
	Description            string           `json:"description"`
	Instructions           string           `json:"instructions"`
	DueDateTime            *DateTime        `json:"dueDateTime"`
	State                  MilestoneState   `json:"state"`
	DepositAmount          Money            `json:"depositAmount"`
	CurrentEscrowAmount    Money            `json:"currentEscrowAmount"`
	FundedAmount           Money            `json:"fundedAmount"`
	Paid                   Money            `json:"paid"`
	Bonus                  Money            `json:"bonus"`
	SubmissionCount        int              `json:"submissionCount"`
	SequenceID             int              `json:"sequenceId"`
	CreatedDateTime        DateTime         `json:"createdDateTime"`
	ModifiedDateTime       DateTime         `json:"modifiedDateTime"`
	CreatedBy              User             `json:"createdBy"`
	ModifiedBy             User             `json:"modifiedBy"`
	SubmissionEvents       []SubmissionEvent `json:"submissionEvents"`
}

// MilestoneState represents the state of a milestone
type MilestoneState string

const (
	MilestoneStateNotFunded  MilestoneState = "NOT_FUNDED"
	MilestoneStateActive     MilestoneState = "ACTIVE"
	MilestoneStateSubmitted  MilestoneState = "SUBMITTED"
	MilestoneStateApproved   MilestoneState = "APPROVED"
	MilestoneStateRejected   MilestoneState = "REJECTED"
	MilestoneStatePaid       MilestoneState = "PAID"
	MilestoneStateCancelled  MilestoneState = "CANCELLED"
)

// SubmissionEvent represents a milestone submission event
type SubmissionEvent struct {
	Submission       *Submission       `json:"submission"`
	SubmissionMessage *SubmissionMessage `json:"submissionMessage"`
	RevisionMessage  *RevisionMessage  `json:"revisionMessage"`
}

// Submission represents a milestone submission
type Submission struct {
	ID               ID       `json:"id"`
	CreatedDateTime  DateTime `json:"createdDateTime"`
	ModifiedDateTime DateTime `json:"modifiedDateTime"`
	Amount           Money    `json:"amount"`
	SequenceID       int      `json:"sequenceId"`
}

// SubmissionMessage represents a submission message
type SubmissionMessage struct {
	CreatedDateTime DateTime `json:"createdDateTime"`
	Message         string   `json:"message"`
}

// RevisionMessage represents a revision message
type RevisionMessage struct {
	CreatedDateTime DateTime `json:"createdDateTime"`
	Message         string   `json:"message"`
}

// CreateMilestoneInput represents input for creating a milestone
type CreateMilestoneInput struct {
	OfferID       string   `json:"offerId"`
	ContractID    string   `json:"contractId"`
	Description   string   `json:"description"`
	Instructions  string   `json:"instruction"`
	DepositAmount string   `json:"depositAmount"`
	DueDate       string   `json:"dueDate"`
	AttachmentIDs []string `json:"attachmentIds,omitempty"`
}

// CreateMilestone creates a new milestone
func (s *ContractsService) CreateMilestone(ctx context.Context, input CreateMilestoneInput) (*Milestone, error) {
	mutation := `
		mutation CreateMilestone(
			$offerId: ID!,
			$contractId: ID!,
			$description: String!,
			$instruction: String!,
			$depositAmount: String!,
			$dueDate: String!,
			$attachmentIds: [ID!]
		) {
			createMilestone(
				input: {
					offerId: $offerId,
					contractId: $contractId,
					description: $description,
					instruction: $instruction,
					depositAmount: $depositAmount,
					dueDate: $dueDate,
					attachmentIds: $attachmentIds
				}
			) {
				id
				description
				instructions
				dueDateTime
				state
				depositAmount {
					rawValue
					currency
					displayValue
				}
				createdDateTime
				sequenceId
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"offerId":       input.OfferID,
			"contractId":    input.ContractID,
			"description":   input.Description,
			"instruction":   input.Instructions,
			"depositAmount": input.DepositAmount,
			"dueDate":       input.DueDate,
			"attachmentIds": input.AttachmentIDs,
		},
	}
	
	var resp struct {
		CreateMilestone Milestone `json:"createMilestone"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.CreateMilestone, nil
}

// EditMilestoneInput represents input for editing a milestone
type EditMilestoneInput struct {
	ID            string   `json:"id"`
	Description   string   `json:"description,omitempty"`
	Instructions  string   `json:"instructions,omitempty"`
	DepositAmount string   `json:"depositAmount,omitempty"`
	DueDate       string   `json:"dueDate,omitempty"`
	Attachments   []string `json:"attachments,omitempty"`
	Message       string   `json:"message,omitempty"`
	SequenceID    int      `json:"sequenceId,omitempty"`
}

// EditMilestone edits an existing milestone
func (s *ContractsService) EditMilestone(ctx context.Context, input EditMilestoneInput) (*Milestone, error) {
	mutation := `
		mutation EditMilestone(
			$id: ID!,
			$description: String,
			$instructions: String,
			$depositAmount: String,
			$dueDate: String,
			$attachments: [ID!],
			$message: String,
			$sequenceId: Int
		) {
			editMilestone(
				input: {
					id: $id,
					description: $description,
					instructions: $instructions,
					depositAmount: $depositAmount,
					dueDate: $dueDate,
					attachments: $attachments,
					message: $message,
					sequenceId: $sequenceId
				}
			) {
				id
				description
				instructions
				dueDateTime
				state
				depositAmount {
					rawValue
					currency
					displayValue
				}
				modifiedDateTime
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"id":            input.ID,
			"description":   input.Description,
			"instructions":  input.Instructions,
			"depositAmount": input.DepositAmount,
			"dueDate":       input.DueDate,
			"attachments":   input.Attachments,
			"message":       input.Message,
			"sequenceId":    input.SequenceID,
		},
	}
	
	var resp struct {
		EditMilestone Milestone `json:"editMilestone"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.EditMilestone, nil
}

// ActivateMilestone activates a milestone
func (s *ContractsService) ActivateMilestone(ctx context.Context, milestoneID string, message string) (*Milestone, error) {
	mutation := `
		mutation ActivateMilestone($id: ID!, $message: String) {
			activateMilestone(input: {id: $id, message: $message}) {
				id
				state
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"id":      milestoneID,
			"message": message,
		},
	}
	
	var resp struct {
		ActivateMilestone Milestone `json:"activateMilestone"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.ActivateMilestone, nil
}

// ApproveMilestoneInput represents input for approving a milestone
type ApproveMilestoneInput struct {
	ID                  string `json:"id"`
	PaidAmount          string `json:"paidAmount,omitempty"`
	BonusAmount         string `json:"bonusAmount,omitempty"`
	PaymentComment      string `json:"paymentComment,omitempty"`
	UnderpaymentReason  string `json:"underpaymentReason,omitempty"`
	NoteToContractor    string `json:"noteToContractor,omitempty"`
}

// ApproveMilestone approves a milestone
func (s *ContractsService) ApproveMilestone(ctx context.Context, input ApproveMilestoneInput) (*Milestone, error) {
	mutation := `
		mutation ApproveMilestone(
			$id: ID!,
			$paidAmount: String,
			$bonusAmount: String,
			$paymentComment: String,
			$underpaymentReason: String,
			$noteToContractor: String
		) {
			approveMilestone(
				input: {
					id: $id,
					paidAmount: $paidAmount,
					bonusAmount: $bonusAmount,
					paymentComment: $paymentComment,
					underpaymentReason: $underpaymentReason,
					noteToContractor: $noteToContractor
				}
			) {
				id
				state
				paid {
					rawValue
					currency
					displayValue
				}
				bonus {
					rawValue
					currency
					displayValue
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"id":                 input.ID,
			"paidAmount":         input.PaidAmount,
			"bonusAmount":        input.BonusAmount,
			"paymentComment":     input.PaymentComment,
			"underpaymentReason": input.UnderpaymentReason,
			"noteToContractor":   input.NoteToContractor,
		},
	}
	
	var resp struct {
		ApproveMilestone Milestone `json:"approveMilestone"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.ApproveMilestone, nil
}

// RejectMilestoneInput represents input for rejecting a milestone
type RejectMilestoneInput struct {
	ID               string `json:"id"`
	NoteToContractor string `json:"noteToContractor"`
}

// RejectMilestone rejects a milestone submission
func (s *ContractsService) RejectMilestone(ctx context.Context, input RejectMilestoneInput) (*Milestone, error) {
	mutation := `
		mutation RejectSubmittedMilestone($id: String, $noteToContractor: String) {
			rejectSubmittedMilestone(
				input: {id: $id, noteToContractor: $noteToContractor}
			) {
				id
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"id":               input.ID,
			"noteToContractor": input.NoteToContractor,
		},
	}
	
	var resp struct {
		RejectSubmittedMilestone struct {
			ID string `json:"id"`
		} `json:"rejectSubmittedMilestone"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	// Get full milestone details
	return s.GetMilestone(ctx, resp.RejectSubmittedMilestone.ID)
}

// DeleteMilestone deletes a milestone
func (s *ContractsService) DeleteMilestone(ctx context.Context, milestoneID string) error {
	mutation := `
		mutation DeleteMilestone($id: ID!) {
			deleteMilestone(input: {id: $id})
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"id": milestoneID,
		},
	}
	
	var resp struct {
		DeleteMilestone bool `json:"deleteMilestone"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.DeleteMilestone {
		return fmt.Errorf("failed to delete milestone")
	}
	
	return nil
}

// GetMilestone retrieves a milestone by ID
func (s *ContractsService) GetMilestone(ctx context.Context, milestoneID string) (*Milestone, error) {
	// This would typically be part of the contract query
	// For now, returning an error as milestone-specific query may not exist
	return nil, fmt.Errorf("GetMilestone not implemented - retrieve via contract")
}

// GetContractMilestones retrieves all milestones for a contract
func (s *ContractsService) GetContractMilestones(ctx context.Context, contractID string) ([]Milestone, error) {
	contract, err := s.GetContract(ctx, contractID)
	if err != nil {
		return nil, err
	}
	
	return contract.Milestones, nil
}