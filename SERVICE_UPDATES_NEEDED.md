# Service Updates Needed for Group Finder

The `internal/groups/service.go` file needs the following methods appended at the end:

## Public Group Operations

```go
// ListPublicGroups - filters groups that are is_open=true and is_active=true
func (s *Service) ListPublicGroups(ctx context.Context, tenantID string, category string, meetingDay string, search string) ([]Group, error)

// GetPublicGroup - gets a single group only if is_open=true
func (s *Service) GetPublicGroup(ctx context.Context, tenantID, groupID string) (*Group, error)
```

## Join Request Operations

```go
// CreateJoinRequest - creates a new join request after validating group is open
func (s *Service) CreateJoinRequest(ctx context.Context, tenantID string, req *JoinRequest) (*JoinRequest, error)

// ListJoinRequests - lists join requests with optional filters (groupID, status)
func (s *Service) ListJoinRequests(ctx context.Context, tenantID string, groupID string, status string) ([]JoinRequest, error)

// GetJoinRequest - gets a single join request
func (s *Service) GetJoinRequest(ctx context.Context, tenantID, requestID string) (*JoinRequest, error)

// UpdateJoinRequestStatus - updates the status of a join request
func (s *Service) UpdateJoinRequestStatus(ctx context.Context, tenantID, requestID, status string) error

// ApproveJoinRequest - approves a join request (future: auto-add to group)
func (s *Service) ApproveJoinRequest(ctx context.Context, tenantID, requestID string) error
```

## Updates to Existing Methods

All existing SELECT queries need to include the new fields:
- `g.category`
- `g.is_open`
- `g.image_url`

Affected methods:
- ListGroups
- GetGroupByID
- GetPersonGroups

All INSERT/UPDATE queries need to include:
- CreateGroup: add category, is_open, image_url to INSERT
- UpdateGroup: add category, is_open, image_url to UPDATE

See the full implementation in the earlier chat context or refer to the handler.go file which already has the correct structure.

## Quick Fix

To complete the implementation, copy the service methods from the successful build earlier in this session, or manually add them following the patterns in handler.go.
