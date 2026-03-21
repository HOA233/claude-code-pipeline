package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ApprovalService manages approval workflows
type ApprovalService struct {
	mu        sync.RWMutex
	requests  map[string]*ApprovalRequest
	workflows map[string]*ApprovalWorkflow
	rules     map[string]*ApprovalRule
}

// ApprovalRequest represents an approval request
type ApprovalRequest struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // deployment, config_change, access_request
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Requester    string                 `json:"requester"`
	TargetID     string                 `json:"target_id"`
	TargetType   string                 `json:"target_type"`
	Priority     ApprovalPriority       `json:"priority"`
	Status       ApprovalStatus         `json:"status"`
	Approvals    []ApprovalDecision     `json:"approvals"`
	RequiredApprovers int               `json:"required_approvers"`
	ApprovedCount int                   `json:"approved_count"`
	RejectedCount int                   `json:"rejected_count"`
	AutoApprove  bool                   `json:"auto_approve"`
	Timeout      time.Duration          `json:"timeout"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ApprovalDecision represents an approval decision
type ApprovalDecision struct {
	ID        string    `json:"id"`
	RequestID string    `json:"request_id"`
	Approver  string    `json:"approver"`
	Action    string    `json:"action"` // approved, rejected, commented
	Comment   string    `json:"comment,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ApprovalWorkflow represents an approval workflow
type ApprovalWorkflow struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        string           `json:"type"`
	Steps       []ApprovalStep   `json:"steps"`
	Enabled     bool             `json:"enabled"`
	CreatedAt   time.Time        `json:"created_at"`
}

// ApprovalStep represents a step in the approval workflow
type ApprovalStep struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Order           int      `json:"order"`
	Approvers       []string `json:"approvers"`
	RequiredCount   int      `json:"required_count"`
	Timeout         time.Duration `json:"timeout"`
	AutoApprove     bool     `json:"auto_approve"`
	AutoReject      bool     `json:"auto_reject"`
}

// ApprovalRule represents an approval rule
type ApprovalRule struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	Conditions      map[string]string `json:"conditions"`
	RequiredApprovers int             `json:"required_approvers"`
	AutoApprove     bool              `json:"auto_approve"`
	Timeout         time.Duration     `json:"timeout"`
	Enabled         bool              `json:"enabled"`
}

// ApprovalPriority represents approval priority
type ApprovalPriority string

const (
	ApprovalPriorityLow    ApprovalPriority = "low"
	ApprovalPriorityNormal ApprovalPriority = "normal"
	ApprovalPriorityHigh   ApprovalPriority = "high"
	ApprovalPriorityUrgent ApprovalPriority = "urgent"
)

// ApprovalStatus represents approval status
type ApprovalStatus string

const (
	ApprovalStatusPending   ApprovalStatus = "pending"
	ApprovalStatusApproved  ApprovalStatus = "approved"
	ApprovalStatusRejected  ApprovalStatus = "rejected"
	ApprovalStatusExpired   ApprovalStatus = "expired"
	ApprovalStatusCancelled ApprovalStatus = "cancelled"
)

// NewApprovalService creates a new approval service
func NewApprovalService() *ApprovalService {
	return &ApprovalService{
		requests:  make(map[string]*ApprovalRequest),
		workflows: make(map[string]*ApprovalWorkflow),
		rules:     make(map[string]*ApprovalRule),
	}
}

// CreateRequest creates a new approval request
func (s *ApprovalService) CreateRequest(ctx context.Context, req *ApprovalRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Requester == "" {
		return fmt.Errorf("requester is required")
	}
	if req.Type == "" {
		return fmt.Errorf("type is required")
	}

	now := time.Now()
	if req.ID == "" {
		req.ID = generateID()
	}
	req.Status = ApprovalStatusPending
	req.CreatedAt = now
	req.UpdatedAt = now
	req.Approvals = make([]ApprovalDecision, 0)
	req.ApprovedCount = 0
	req.RejectedCount = 0

	// Set default required approvers
	if req.RequiredApprovers == 0 {
		req.RequiredApprovers = 1
	}

	// Set expiry if timeout specified
	if req.Timeout > 0 {
		expiresAt := now.Add(req.Timeout)
		req.ExpiresAt = &expiresAt
	}

	// Check for auto-approve
	if req.AutoApprove {
		req.Status = ApprovalStatusApproved
		completedAt := now
		req.CompletedAt = &completedAt
	}

	s.requests[req.ID] = req

	return nil
}

// GetRequest gets an approval request by ID
func (s *ApprovalService) GetRequest(id string) (*ApprovalRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.requests[id]
	if !ok {
		return nil, fmt.Errorf("approval request not found: %s", id)
	}
	return req, nil
}

// Approve approves an approval request
func (s *ApprovalService) Approve(requestID, approver, comment string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.requests[requestID]
	if !ok {
		return fmt.Errorf("approval request not found: %s", requestID)
	}

	if req.Status != ApprovalStatusPending {
		return fmt.Errorf("request is not pending")
	}

	// Check if already approved by this approver
	for _, a := range req.Approvals {
		if a.Approver == approver && a.Action == "approved" {
			return fmt.Errorf("already approved by this approver")
		}
	}

	now := time.Now()
	decision := ApprovalDecision{
		ID:        generateID(),
		RequestID: requestID,
		Approver:  approver,
		Action:    "approved",
		Comment:   comment,
		Timestamp: now,
	}

	req.Approvals = append(req.Approvals, decision)
	req.ApprovedCount++
	req.UpdatedAt = now

	// Check if we have enough approvals
	if req.ApprovedCount >= req.RequiredApprovers {
		req.Status = ApprovalStatusApproved
		req.CompletedAt = &now
	}

	return nil
}

// Reject rejects an approval request
func (s *ApprovalService) Reject(requestID, approver, comment string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.requests[requestID]
	if !ok {
		return fmt.Errorf("approval request not found: %s", requestID)
	}

	if req.Status != ApprovalStatusPending {
		return fmt.Errorf("request is not pending")
	}

	now := time.Now()
	decision := ApprovalDecision{
		ID:        generateID(),
		RequestID: requestID,
		Approver:  approver,
		Action:    "rejected",
		Comment:   comment,
		Timestamp: now,
	}

	req.Approvals = append(req.Approvals, decision)
	req.RejectedCount++
	req.Status = ApprovalStatusRejected
	req.CompletedAt = &now
	req.UpdatedAt = now

	return nil
}

// Comment adds a comment to an approval request
func (s *ApprovalService) Comment(requestID, commenter, comment string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.requests[requestID]
	if !ok {
		return fmt.Errorf("approval request not found: %s", requestID)
	}

	decision := ApprovalDecision{
		ID:        generateID(),
		RequestID: requestID,
		Approver:  commenter,
		Action:    "commented",
		Comment:   comment,
		Timestamp: time.Now(),
	}

	req.Approvals = append(req.Approvals, decision)
	req.UpdatedAt = time.Now()

	return nil
}

// Cancel cancels an approval request
func (s *ApprovalService) Cancel(requestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.requests[requestID]
	if !ok {
		return fmt.Errorf("approval request not found: %s", requestID)
	}

	if req.Status != ApprovalStatusPending {
		return fmt.Errorf("can only cancel pending requests")
	}

	now := time.Now()
	req.Status = ApprovalStatusCancelled
	req.CompletedAt = &now
	req.UpdatedAt = now

	return nil
}

// ListRequests lists approval requests with filters
func (s *ApprovalService) ListRequests(status ApprovalStatus, requester string) []*ApprovalRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*ApprovalRequest
	for _, req := range s.requests {
		if status != "" && req.Status != status {
			continue
		}
		if requester != "" && req.Requester != requester {
			continue
		}
		results = append(results, req)
	}
	return results
}

// GetPendingApprovals gets pending approvals for an approver
func (s *ApprovalService) GetPendingApprovals(approver string) []*ApprovalRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*ApprovalRequest
	for _, req := range s.requests {
		if req.Status != ApprovalStatusPending {
			continue
		}

		// Check if approver has already acted
		acted := false
		for _, a := range req.Approvals {
			if a.Approver == approver {
				acted = true
				break
			}
		}

		if !acted {
			results = append(results, req)
		}
	}
	return results
}

// CreateWorkflow creates an approval workflow
func (s *ApprovalService) CreateWorkflow(workflow *ApprovalWorkflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workflow.Name == "" {
		return fmt.Errorf("name is required")
	}

	if workflow.ID == "" {
		workflow.ID = generateID()
	}
	workflow.CreatedAt = time.Now()

	s.workflows[workflow.ID] = workflow

	return nil
}

// GetWorkflow gets an approval workflow
func (s *ApprovalService) GetWorkflow(id string) (*ApprovalWorkflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workflow, ok := s.workflows[id]
	if !ok {
		return nil, fmt.Errorf("workflow not found: %s", id)
	}
	return workflow, nil
}

// CreateRule creates an approval rule
func (s *ApprovalService) CreateRule(rule *ApprovalRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rule.Name == "" {
		return fmt.Errorf("name is required")
	}

	if rule.ID == "" {
		rule.ID = generateID()
	}

	s.rules[rule.ID] = rule

	return nil
}

// GetRule gets an approval rule
func (s *ApprovalService) GetRule(id string) (*ApprovalRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rule, ok := s.rules[id]
	if !ok {
		return nil, fmt.Errorf("rule not found: %s", id)
	}
	return rule, nil
}

// CheckExpiredRequests checks and marks expired requests
func (s *ApprovalService) CheckExpiredRequests() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expired := 0

	for _, req := range s.requests {
		if req.Status == ApprovalStatusPending && req.ExpiresAt != nil {
			if now.After(*req.ExpiresAt) {
				req.Status = ApprovalStatusExpired
				req.CompletedAt = &now
				req.UpdatedAt = now
				expired++
			}
		}
	}

	return expired
}

// GetStats gets approval statistics
func (s *ApprovalService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]int{
		"total":      len(s.requests),
		"pending":    0,
		"approved":   0,
		"rejected":   0,
		"expired":    0,
		"cancelled":  0,
	}

	for _, req := range s.requests {
		switch req.Status {
		case ApprovalStatusPending:
			stats["pending"]++
		case ApprovalStatusApproved:
			stats["approved"]++
		case ApprovalStatusRejected:
			stats["rejected"]++
		case ApprovalStatusExpired:
			stats["expired"]++
		case ApprovalStatusCancelled:
			stats["cancelled"]++
		}
	}

	return map[string]interface{}{
		"requests": stats,
		"workflows": len(s.workflows),
		"rules":     len(s.rules),
	}
}

// ToJSON serializes request to JSON
func (r *ApprovalRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSON serializes workflow to JSON
func (w *ApprovalWorkflow) ToJSON() ([]byte, error) {
	return json.Marshal(w)
}