package tests

import (
	"context"
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Approval Service Tests

func TestApprovalService_New(t *testing.T) {
	as := service.NewApprovalService()
	if as == nil {
		t.Fatal("Expected non-nil approval service")
	}
}

func TestApprovalService_CreateRequest(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Type:     "deployment",
		Title:    "Deploy to Production",
		Requester: "user-1",
	}

	err := as.CreateRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if req.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if req.Status != service.ApprovalStatusPending {
		t.Errorf("Expected status pending, got %s", req.Status)
	}
}

func TestApprovalService_CreateRequest_MissingRequester(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Type:  "deployment",
		Title: "Deploy",
	}

	err := as.CreateRequest(context.Background(), req)
	if err == nil {
		t.Error("Expected error for missing requester")
	}
}

func TestApprovalService_CreateRequest_MissingType(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Title:     "Deploy",
		Requester: "user-1",
	}

	err := as.CreateRequest(context.Background(), req)
	if err == nil {
		t.Error("Expected error for missing type")
	}
}

func TestApprovalService_GetRequest(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Type:      "config_change",
		Title:     "Change Config",
		Requester: "user-get",
	}
	as.CreateRequest(context.Background(), req)

	retrieved, err := as.GetRequest(req.ID)
	if err != nil {
		t.Fatalf("Failed to get request: %v", err)
	}

	if retrieved.Title != "Change Config" {
		t.Error("Request title mismatch")
	}
}

func TestApprovalService_GetRequest_NotFound(t *testing.T) {
	as := service.NewApprovalService()

	_, err := as.GetRequest("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent request")
	}
}

func TestApprovalService_Approve(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Type:             "deployment",
		Title:            "Deploy",
		Requester:        "user-approve",
		RequiredApprovers: 1,
	}
	as.CreateRequest(context.Background(), req)

	err := as.Approve(req.ID, "approver-1", "Looks good")
	if err != nil {
		t.Fatalf("Failed to approve: %v", err)
	}

	retrieved, _ := as.GetRequest(req.ID)
	if retrieved.Status != service.ApprovalStatusApproved {
		t.Error("Expected status to be approved")
	}

	if retrieved.ApprovedCount != 1 {
		t.Errorf("Expected 1 approval, got %d", retrieved.ApprovedCount)
	}
}

func TestApprovalService_Approve_MultipleApprovers(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Type:              "deployment",
		Title:             "Deploy",
		Requester:         "user-multi",
		RequiredApprovers: 2,
	}
	as.CreateRequest(context.Background(), req)

	// First approval
	as.Approve(req.ID, "approver-1", "OK")

	retrieved, _ := as.GetRequest(req.ID)
	if retrieved.Status == service.ApprovalStatusApproved {
		t.Error("Should not be approved yet")
	}

	// Second approval
	as.Approve(req.ID, "approver-2", "Approved")

	retrieved, _ = as.GetRequest(req.ID)
	if retrieved.Status != service.ApprovalStatusApproved {
		t.Error("Expected status to be approved after 2 approvals")
	}
}

func TestApprovalService_Reject(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Type:      "deployment",
		Title:     "Deploy",
		Requester: "user-reject",
	}
	as.CreateRequest(context.Background(), req)

	err := as.Reject(req.ID, "approver-1", "Not ready")
	if err != nil {
		t.Fatalf("Failed to reject: %v", err)
	}

	retrieved, _ := as.GetRequest(req.ID)
	if retrieved.Status != service.ApprovalStatusRejected {
		t.Error("Expected status to be rejected")
	}
}

func TestApprovalService_Comment(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Type:      "deployment",
		Title:     "Deploy",
		Requester: "user-comment",
	}
	as.CreateRequest(context.Background(), req)

	err := as.Comment(req.ID, "commenter-1", "Please review the changes")
	if err != nil {
		t.Fatalf("Failed to comment: %v", err)
	}

	retrieved, _ := as.GetRequest(req.ID)
	if len(retrieved.Approvals) != 1 {
		t.Error("Expected one approval entry (comment)")
	}
}

func TestApprovalService_Cancel(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Type:      "deployment",
		Title:     "Deploy",
		Requester: "user-cancel",
	}
	as.CreateRequest(context.Background(), req)

	err := as.Cancel(req.ID)
	if err != nil {
		t.Fatalf("Failed to cancel: %v", err)
	}

	retrieved, _ := as.GetRequest(req.ID)
	if retrieved.Status != service.ApprovalStatusCancelled {
		t.Error("Expected status to be cancelled")
	}
}

func TestApprovalService_ListRequests(t *testing.T) {
	as := service.NewApprovalService()

	as.CreateRequest(context.Background(), &service.ApprovalRequest{
		Type:      "deployment",
		Title:     "Deploy 1",
		Requester: "user-list",
	})

	as.CreateRequest(context.Background(), &service.ApprovalRequest{
		Type:      "config_change",
		Title:     "Config 1",
		Requester: "user-list",
	})

	as.Approve("1", "approver", "OK") // This won't work, just for test

	requests := as.ListRequests(service.ApprovalStatusPending, "user-list")
	if len(requests) < 2 {
		t.Errorf("Expected at least 2 pending requests, got %d", len(requests))
	}
}

func TestApprovalService_GetPendingApprovals(t *testing.T) {
	as := service.NewApprovalService()

	as.CreateRequest(context.Background(), &service.ApprovalRequest{
		Type:      "deployment",
		Title:     "Deploy",
		Requester: "user-pending",
	})

	pending := as.GetPendingApprovals("approver-1")
	if len(pending) < 1 {
		t.Error("Expected at least 1 pending approval")
	}
}

func TestApprovalService_CreateWorkflow(t *testing.T) {
	as := service.NewApprovalService()

	workflow := &service.ApprovalWorkflow{
		Name: "Deployment Approval",
		Type: "deployment",
		Steps: []service.ApprovalStep{
			{Name: "Tech Lead", RequiredCount: 1},
			{Name: "Manager", RequiredCount: 1},
		},
	}

	err := as.CreateWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	if workflow.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestApprovalService_CreateRule(t *testing.T) {
	as := service.NewApprovalService()

	rule := &service.ApprovalRule{
		Name:              "Production Deployments",
		Type:              "deployment",
		RequiredApprovers: 2,
	}

	err := as.CreateRule(rule)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	if rule.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestApprovalService_AutoApprove(t *testing.T) {
	as := service.NewApprovalService()

	req := &service.ApprovalRequest{
		Type:        "deployment",
		Title:       "Auto Deploy",
		Requester:   "user-auto",
		AutoApprove: true,
	}

	as.CreateRequest(context.Background(), req)

	if req.Status != service.ApprovalStatusApproved {
		t.Error("Expected status to be auto-approved")
	}
}

func TestApprovalService_CheckExpiredRequests(t *testing.T) {
	as := service.NewApprovalService()

	// Create request with very short timeout
	req := &service.ApprovalRequest{
		Type:    "deployment",
		Title:   "Expiring",
		Requester: "user-expire",
		Timeout:  1 * time.Nanosecond,
	}
	as.CreateRequest(context.Background(), req)

	// Wait a moment
	time.Sleep(10 * time.Millisecond)

	expired := as.CheckExpiredRequests()
	if expired < 1 {
		t.Error("Expected at least 1 expired request")
	}
}

func TestApprovalService_GetStats(t *testing.T) {
	as := service.NewApprovalService()

	as.CreateRequest(context.Background(), &service.ApprovalRequest{
		Type:      "deployment",
		Title:     "Stats",
		Requester: "user-stats",
	})

	stats := as.GetStats()

	if stats == nil {
		t.Fatal("Expected stats")
	}

	requests := stats["requests"].(map[string]int)
	if requests["total"] < 1 {
		t.Error("Expected at least 1 total request")
	}
}

func TestApprovalService_RequestToJSON(t *testing.T) {
	req := &service.ApprovalRequest{
		ID:        "req-1",
		Type:      "deployment",
		Title:     "Deploy",
		Requester: "user-1",
		Status:    service.ApprovalStatusPending,
		CreatedAt: time.Now(),
	}

	data, err := req.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestApprovalService_Priorities(t *testing.T) {
	priorities := []service.ApprovalPriority{
		service.ApprovalPriorityLow,
		service.ApprovalPriorityNormal,
		service.ApprovalPriorityHigh,
		service.ApprovalPriorityUrgent,
	}

	as := service.NewApprovalService()

	for i, priority := range priorities {
		req := &service.ApprovalRequest{
			Type:      "test",
			Title:     string(priority),
			Requester: "user-priority",
			Priority:  priority,
		}

		err := as.CreateRequest(context.Background(), req)
		if err != nil {
			t.Errorf("Failed to create request with priority %s: %v", priority, err)
		}
		_ = i
	}
}