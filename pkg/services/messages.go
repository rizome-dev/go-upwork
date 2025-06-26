package services

import (
	"github.com/rizome-dev/go-upwork/pkg/models"
	"context"
	"fmt"
)

// MessagesService handles messaging-related API operations
type MessagesService struct {
	client *BaseClient
}

// NewMessagesService creates a new messages service
func NewMessagesService(client *BaseClient) *MessagesService {
	return &MessagesService{client: client}
}

// Room represents a chat room
type Room struct {
	ID                 ID           `json:"id"`
	RoomName           string       `json:"roomName"`
	RoomType           RoomType     `json:"roomType"`
	Topic              string       `json:"topic"`
	NumUnread          int          `json:"numUnread"`
	NumUnreadMentions  int          `json:"numUnreadMentions"`
	NumUsers           int          `json:"numUsers"`
	Favorite           bool         `json:"favorite"`
	ReadOnly           bool         `json:"readOnly"`
	Hidden             bool         `json:"hidden"`
	Public             bool         `json:"public"`
	LastVisitedDateTime *DateTime    `json:"lastVisitedDateTime"`
	LastReadDateTime    *DateTime    `json:"lastReadDateTime"`
	CreatedAtDateTime   DateTime     `json:"createdAtDateTime"`
	Organization        Organization `json:"organization"`
	RoomUsers          []RoomUser   `json:"roomUsers"`
	Creator            *RoomUser    `json:"creator"`
	Owner              *RoomUser    `json:"owner"`
	LatestStory        *Story       `json:"latestStory"`
}

// RoomType represents the type of room
type RoomType string

const (
	RoomTypeGroup       RoomType = "GROUP"
	RoomTypeOneOnOne    RoomType = "ONE_ON_ONE"
	RoomTypeInterview   RoomType = "INTERVIEW"
	RoomTypeContract    RoomType = "CONTRACT"
	RoomTypePublic      RoomType = "PUBLIC"
)

// RoomUser represents a user in a room
type RoomUser struct {
	User         User         `json:"user"`
	Organization Organization `json:"organization"`
	Role         string       `json:"role"`
}

// Story represents a message/story in a room
type Story struct {
	ID               ID       `json:"id"`
	CreatedDateTime  DateTime `json:"createdDateTime"`
	UpdatedDateTime  DateTime `json:"updatedDateTime"`
	User             User     `json:"user"`
	Message          string   `json:"message"`
	Organization     Organization `json:"organization"`
	RoomStoryNote    *RoomStoryNote `json:"roomStoryNote"`
}

// RoomStoryNote represents a note on a story
type RoomStoryNote struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// RoomFilter represents room filtering options
type RoomFilter struct {
	RoomType               RoomType `json:"roomType_eq,omitempty"`
	RoomPrivacy            string   `json:"roomPrivacy_eq,omitempty"`
	Subscribed             bool     `json:"subscribed_eq,omitempty"`
	ActiveSince            string   `json:"activeSince_eq,omitempty"`
	IncludeFavorites       bool     `json:"includeFavorites_eq,omitempty"`
	IncludeUnreadIfActive  bool     `json:"includeUnreadIfActive_eq,omitempty"`
	UnreadRoomsOnly        bool     `json:"unreadRoomsOnly_eq,omitempty"`
	IncludeHidden          bool     `json:"includeHidden_eq,omitempty"`
	ObjectReferenceID      string   `json:"objectReferenceId_eq,omitempty"`
	RoomCategory           string   `json:"roomCategory_eq,omitempty"`
}

// RoomList represents a paginated list of rooms
type RoomList struct {
	TotalCount int        `json:"totalCount"`
	PageInfo   PageInfo   `json:"pageInfo"`
	Edges      []RoomEdge `json:"edges"`
}

// RoomEdge represents a room edge in pagination
type RoomEdge struct {
	Cursor string `json:"cursor"`
	Node   Room   `json:"node"`
}

// ListRooms returns a list of rooms
func (s *MessagesService) ListRooms(ctx context.Context, filter *RoomFilter, pagination *PaginationInput, sortOrder SortOrder) (*RoomList, error) {
	query := `
		query ListRooms($filter: RoomFilter, $pagination: Pagination, $sortOrder: SortOrder) {
			roomList(filter: $filter, pagination: $pagination, sortOrder: $sortOrder) {
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
						roomName
						roomType
						topic
						numUnread
						numUnreadMentions
						numUsers
						favorite
						createdAtDateTime
						latestStory {
							createdDateTime
							updatedDateTime
						}
						organization {
							id
							legacyId
						}
						roomUsers {
							user {
								id
								name
							}
							role
						}
					}
				}
			}
		}
	`
	
	variables := map[string]interface{}{}
	if filter != nil {
		variables["filter"] = filter
	}
	if pagination != nil {
		variables["pagination"] = pagination
	}
	if sortOrder != "" {
		variables["sortOrder"] = sortOrder
	}
	
	req := &GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	
	var resp struct {
		RoomList RoomList `json:"roomList"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.RoomList, nil
}

// GetRoom returns a specific room by ID
func (s *MessagesService) GetRoom(ctx context.Context, roomID string) (*Room, error) {
	query := `
		query GetRoom($roomId: ID!) {
			room(id: $roomId) {
				id
				roomName
				roomType
				topic
				lastVisitedDateTime
				lastReadDateTime
				favorite
				readOnly
				hidden
				public
				organization {
					id
				}
				roomUsers {
					user {
						id
						name
					}
					organization {
						id
					}
					role
				}
				creator {
					user {
						id
						name
					}
				}
				owner {
					user {
						id
					}
				}
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"roomId": roomID,
		},
	}
	
	var resp struct {
		Room Room `json:"room"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.Room, nil
}

// CreateRoomInput represents input for creating a room
type CreateRoomInput struct {
	RoomName string     `json:"roomName"`
	Topic    string     `json:"topic"`
	RoomType RoomType   `json:"roomType"`
	Users    []RoomUserInput `json:"users"`
}

// RoomUserInput represents input for room users
type RoomUserInput struct {
	UserID         string `json:"userId"`
	OrganizationID string `json:"organizationId"`
}

// CreateRoom creates a new room
func (s *MessagesService) CreateRoom(ctx context.Context, input CreateRoomInput) (*Room, error) {
	mutation := `
		mutation CreateRoom($input: RoomCreateInputV2!) {
			createRoomV2(input: $input) {
				id
				roomName
				roomType
				topic
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
		CreateRoomV2 Room `json:"createRoomV2"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.CreateRoomV2, nil
}

// CreateStoryInput represents input for creating a story/message
type CreateStoryInput struct {
	RoomID  string `json:"roomId"`
	Message string `json:"message"`
}

// SendMessage sends a message to a room
func (s *MessagesService) SendMessage(ctx context.Context, input CreateStoryInput) (*Story, error) {
	mutation := `
		mutation SendMessage($input: RoomStoryCreateInputV2!) {
			createRoomStoryV2(input: $input) {
				id
				createdDateTime
				updatedDateTime
				message
				user {
					id
					name
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
		CreateRoomStoryV2 Story `json:"createRoomStoryV2"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.CreateRoomStoryV2, nil
}

// GetRoomStories returns stories/messages from a room
func (s *MessagesService) GetRoomStories(ctx context.Context, roomID string, pagination *PaginationInput) ([]Story, error) {
	query := `
		query GetRoomStories($roomId: ID!, $pagination: Pagination) {
			roomStories(filter: {roomId_eq: $roomId}, pagination: $pagination) {
				totalCount
				edges {
					node {
						id
						message
						createdDateTime
						updatedDateTime
						user {
							id
							name
						}
					}
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"roomId": roomID,
	}
	if pagination != nil {
		variables["pagination"] = pagination
	}
	
	req := &GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	
	var resp struct {
		RoomStories struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node Story `json:"node"`
			} `json:"edges"`
		} `json:"roomStories"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	stories := make([]Story, 0, len(resp.RoomStories.Edges))
	for _, edge := range resp.RoomStories.Edges {
		stories = append(stories, edge.Node)
	}
	
	return stories, nil
}

// UpdateRoomInput represents input for updating a room
type UpdateRoomInput struct {
	RoomID string `json:"roomId"`
	Topic  string `json:"topic,omitempty"`
	Name   string `json:"name,omitempty"`
}

// UpdateRoom updates room settings
func (s *MessagesService) UpdateRoom(ctx context.Context, input UpdateRoomInput) (*Room, error) {
	mutation := `
		mutation UpdateRoom($roomId: ID!, $topic: String!) {
			updateRoom(input: {roomId: $roomId, topic: $topic}) {
				id
				roomName
				topic
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"roomId": input.RoomID,
			"topic":  input.Topic,
		},
	}
	
	var resp struct {
		UpdateRoom Room `json:"updateRoom"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.UpdateRoom, nil
}

// ArchiveRoom archives a room
func (s *MessagesService) ArchiveRoom(ctx context.Context, roomID string) (*Room, error) {
	mutation := `
		mutation ArchiveRoom($roomId: ID!) {
			archiveRoom(roomId: $roomId) {
				id
				hidden
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"roomId": roomID,
		},
	}
	
	var resp struct {
		ArchiveRoom Room `json:"archiveRoom"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.ArchiveRoom, nil
}

// GetRoomByOfferID returns a room associated with an offer
func (s *MessagesService) GetRoomByOfferID(ctx context.Context, offerID string) (*Room, error) {
	query := `
		query GetOfferRoom($offerId: ID!) {
			offerRoom(id: $offerId) {
				id
				roomName
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"offerId": offerID,
		},
	}
	
	var resp struct {
		OfferRoom Room `json:"offerRoom"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.OfferRoom, nil
}

// GetRoomByContractID returns a room associated with a contract
func (s *MessagesService) GetRoomByContractID(ctx context.Context, contractID string) (*Room, error) {
	query := `
		query GetContractRoom($contractId: ID!) {
			contractRoom(id: $contractId) {
				id
				roomName
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"contractId": contractID,
		},
	}
	
	var resp struct {
		ContractRoom Room `json:"contractRoom"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.ContractRoom, nil
}

// GetRoomByProposalID returns a room associated with a proposal
func (s *MessagesService) GetRoomByProposalID(ctx context.Context, proposalID string) (*Room, error) {
	query := `
		query GetProposalRoom($vendorProposalId: ID!) {
			proposalRoom(id: $vendorProposalId) {
				id
				roomName
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"vendorProposalId": proposalID,
		},
	}
	
	var resp struct {
		ProposalRoom Room `json:"proposalRoom"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, err
	}
	
	return &resp.ProposalRoom, nil
}

// AddUserToRoom adds a user to a room
func (s *MessagesService) AddUserToRoom(ctx context.Context, roomID string, userID string) error {
	mutation := `
		mutation AddUserToRoom($roomId: ID!, $userId: ID!) {
			addUserToRoom(roomId: $roomId, userId: $userId) {
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"roomId": roomID,
			"userId": userID,
		},
	}
	
	var resp struct {
		AddUserToRoom struct {
			Success bool `json:"success"`
		} `json:"addUserToRoom"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.AddUserToRoom.Success {
		return fmt.Errorf("failed to add user to room")
	}
	
	return nil
}

// RemoveUserFromRoom removes a user from a room
func (s *MessagesService) RemoveUserFromRoom(ctx context.Context, roomID string, userID string) error {
	mutation := `
		mutation RemoveUserFromRoom($roomId: ID!, $userId: ID!) {
			removeUserFromRoom(roomId: $roomId, userId: $userId) {
				success
			}
		}
	`
	
	req := &GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"roomId": roomID,
			"userId": userID,
		},
	}
	
	var resp struct {
		RemoveUserFromRoom struct {
			Success bool `json:"success"`
		} `json:"removeUserFromRoom"`
	}
	
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return err
	}
	
	if !resp.RemoveUserFromRoom.Success {
		return fmt.Errorf("failed to remove user from room")
	}
	
	return nil
}