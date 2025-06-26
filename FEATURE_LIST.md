# Upwork Go SDK - Core Features List

## 1. Authentication & Authorization
- OAuth 2.0 implementation with multiple grant types:
  - Authorization Code Grant
  - Implicit Grant  
  - Client Credentials Grant (Enterprise only)
  - Refresh Token Grant
- Service Account support
- API Key management
- Scopes and permissions management
- Required X-Upwork-API-TenantId header support

## 2. Messaging API
### Queries
- roomList - List all rooms
- room - Get specific room by ID
- roomStories - Get messages from a room
- offerRoom - Get room by offer ID
- proposalRoom - Get room by application/proposal ID
- contractRoom - Get room by contract ID
- oneOnOneRoom - Get 1-on-1 room

### Mutations  
- createRoomV2 - Create new room
- createRoomStoryV2 - Send message to room
- updateRoomV2 - Update room settings
- archiveRoom - Archive/hide room
- addUserToRoom - Add user to room
- removeUserFromRoom - Remove user from room
- removeRoom - Delete room
- removeRoomStory - Delete message

## 3. Contracts & Offers API
### Queries
- contract - Get contract by ID
- contractList - List contracts
- contractDetails - Get detailed contract info
- contractByTerm - Find contracts by terms
- offer - Get offer by ID
- offersByIds - Get multiple offers
- offersByAttribute - Search offers
- vendorContracts - Get freelancer's contracts

### Mutations
- createOffer - Create new offer
- withdrawOffer - Withdraw offer
- endContractByClient - End contract (client)
- endContractByFreelancer - End contract (freelancer)
- pauseContract - Pause contract
- restartContract - Restart contract
- updateContractHourlyLimit - Update hourly limit

## 4. Payments & Milestones API
### Queries
- transactionHistory - Get payment history

### Mutations
- sendCustomPayment - Make custom payment
- createMilestoneV2 - Create milestone
- editMilestone - Edit milestone
- activateMilestone - Activate milestone
- approveMilestone - Approve milestone
- deleteMilestone - Delete milestone
- rejectSubmittedMilestone - Reject milestone submission

## 5. Job Posting API
### Queries
- jobPosting - Get job posting
- marketplaceJobPosting - Get public job
- marketplaceJobPostings - Search jobs
- marketplaceJobPostingsSearch - Advanced job search
- marketplaceJobPostingsContents - Get job contents
- jobsFeaturePredictions - Get job predictions

### Mutations
- createJobPosting - Create job posting
- updateJobPosting - Update job posting

## 6. User & Organization Management API
### Queries
- user - Get current user
- userDetails - Get user details by ID
- userIdsByEmail - Find users by email
- organization - Get organization info
- company - Get company info
- companySelector - List user's companies
- staffsByPersonId - Get staff by person

### Mutations
- createOrganization - Create organization
- updateOrganization - Update organization
- inviteToTeam - Invite to team

## 7. Freelancer Profile API
### Queries
- freelancerProfileByProfileKey - Get profile
- freelancerVisibility - Get visibility settings
- talentProfileByProfileKey - Get talent profile
- freelancerProfileSearchRecords - Search profiles
- search - General search endpoint

### Mutations
- addFreelancerEmploymentRecord - Add employment
- removeFreelancerEmploymentRecord - Remove employment
- updateFreelancerEmploymentRecord - Update employment
- addFreelancerOtherExperience - Add experience
- removeFreelancerOtherExperience - Remove experience
- updateFreelancerOtherExperience - Update experience
- addFreelancerLanguage - Add language
- removeFreelancerLanguage - Remove language
- updateFreelancerAvailability - Update availability

## 8. Work Diary & Time Tracking API
### Queries
- workDiaryCompany - Get company work diary
- workDiaryContract - Get contract work diary
- workDiaryCellActivities - Get cell activities
- workDays - Get work days
- snapshotsByContractId - Get snapshots
- contractTimeReport - Get time reports
- timeReport - Get time report

## 9. Activities & Tasks API
### Queries
- teamActivities - List team activities
- talentCloudTasks - Get talent cloud tasks

### Mutations
- addTeamActivity - Create activity
- updateTeamActivity - Update activity
- archiveTeamActivity - Archive activity
- unarchiveTeamActivity - Unarchive activity
- assignTeamActivityToTheContract - Assign to contract

## 10. Proposals API
### Queries
- clientProposal - Get client proposal
- clientProposals - List client proposals
- vendorProposal - Get vendor proposal
- vendorProposals - List vendor proposals
- proposalMetadata - Get proposal metadata

### Mutations
- declineClientProposal - Decline proposal
- hideClientProposal - Hide proposal
- markClientProposalAsRead - Mark as read
- messageClientProposal - Message about proposal
- shortlistClientProposal - Shortlist proposal
- createDirectUploadLinkForJAClientProposal - Create upload link

## 11. Metadata & Reference Data API
### Queries
- ontologyCategories - List categories
- ontologySkills - List skills
- ontologyOccupations - List occupations
- countries - List countries
- languages - List languages
- reasons - List reasons
- regions - List regions
- timeZones - List time zones

## 12. Reports & Analytics API
### Queries
- transactionHistory - Financial transactions
- contractTimeReport - Time reports
- accountingEntity - Accounting entities

## 13. Workflow API
### Queries
- workflowView - Get workflow view

### Mutations
- updateWorkflowTask - Update workflow task
- confirmFiles - Confirm files

## 14. Other Features
- GraphQL Subscriptions support
- Error handling with detailed error types
- Pagination support
- Filtering and sorting
- File upload support
- Real-time notifications (via subscriptions)

## API Configuration
- Base URL: https://api.upwork.com/graphql
- Authentication: OAuth 2.0
- Rate Limits: 300 requests/minute per IP
- Request size limits
- Caching restrictions (24 hours max)