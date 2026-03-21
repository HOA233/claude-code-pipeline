package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Notification Service Tests

func TestNotificationService_New(t *testing.T) {
	ns := service.NewNotificationService()
	if ns == nil {
		t.Fatal("Expected non-nil notification service")
	}
}

func TestNotificationService_Send(t *testing.T) {
	ns := service.NewNotificationService()

	notification := &service.Notification{
		Type:      service.NotificationTypeAlert,
		Title:     "Test Notification",
		Message:   "This is a test",
		Recipient: "user-1",
		TenantID:  "tenant-1",
	}

	err := ns.Send(notification)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	if notification.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if notification.Status != service.NotificationStatusSent {
		t.Errorf("Expected status sent, got %s", notification.Status)
	}
}

func TestNotificationService_Send_MissingRecipient(t *testing.T) {
	ns := service.NewNotificationService()

	notification := &service.Notification{
		Title:    "No Recipient",
		Message:  "Test",
		TenantID: "tenant-1",
	}

	err := ns.Send(notification)
	if err == nil {
		t.Error("Expected error for missing recipient")
	}
}

func TestNotificationService_Get(t *testing.T) {
	ns := service.NewNotificationService()

	notification := &service.Notification{
		Type:      service.NotificationTypeInfo,
		Title:     "Get Test",
		Message:   "Test message",
		Recipient: "user-get",
		TenantID:  "tenant-get",
	}
	ns.Send(notification)

	retrieved, err := ns.Get(notification.ID)
	if err != nil {
		t.Fatalf("Failed to get notification: %v", err)
	}

	if retrieved.Title != "Get Test" {
		t.Error("Notification title mismatch")
	}
}

func TestNotificationService_Get_NotFound(t *testing.T) {
	ns := service.NewNotificationService()

	_, err := ns.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent notification")
	}
}

func TestNotificationService_ListByRecipient(t *testing.T) {
	ns := service.NewNotificationService()

	ns.Send(&service.Notification{
		Type:      service.NotificationTypeAlert,
		Title:     "List 1",
		Message:   "Test",
		Recipient: "user-list",
		TenantID:  "tenant-list",
	})

	ns.Send(&service.Notification{
		Type:      service.NotificationTypeInfo,
		Title:     "List 2",
		Message:   "Test",
		Recipient: "user-list",
		TenantID:  "tenant-list",
	})

	notifications := ns.ListByRecipient("user-list")
	if len(notifications) < 2 {
		t.Errorf("Expected at least 2 notifications, got %d", len(notifications))
	}
}

func TestNotificationService_ListByTenant(t *testing.T) {
	ns := service.NewNotificationService()

	ns.Send(&service.Notification{
		Type:      service.NotificationTypeWarning,
		Title:     "Tenant List",
		Message:   "Test",
		Recipient: "user-1",
		TenantID:  "tenant-tenantlist",
	})

	notifications := ns.ListByTenant("tenant-tenantlist")
	if len(notifications) < 1 {
		t.Errorf("Expected at least 1 notification, got %d", len(notifications))
	}
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	ns := service.NewNotificationService()

	notification := &service.Notification{
		Type:      service.NotificationTypeAlert,
		Title:     "Mark Read",
		Message:   "Test",
		Recipient: "user-read",
		TenantID:  "tenant-read",
	}
	ns.Send(notification)

	err := ns.MarkAsRead(notification.ID)
	if err != nil {
		t.Fatalf("Failed to mark as read: %v", err)
	}

	retrieved, _ := ns.Get(notification.ID)
	if !retrieved.Read {
		t.Error("Expected notification to be marked as read")
	}
}

func TestNotificationService_MarkAllAsRead(t *testing.T) {
	ns := service.NewNotificationService()

	ns.Send(&service.Notification{
		Type:      service.NotificationTypeAlert,
		Title:     "Mark All 1",
		Message:   "Test",
		Recipient: "user-markall",
		TenantID:  "tenant-markall",
	})

	ns.Send(&service.Notification{
		Type:      service.NotificationTypeInfo,
		Title:     "Mark All 2",
		Message:   "Test",
		Recipient: "user-markall",
		TenantID:  "tenant-markall",
	})

	count := ns.MarkAllAsRead("user-markall")
	if count < 2 {
		t.Errorf("Expected at least 2 notifications marked, got %d", count)
	}
}

func TestNotificationService_Delete(t *testing.T) {
	ns := service.NewNotificationService()

	notification := &service.Notification{
		Type:      service.NotificationTypeAlert,
		Title:     "Delete Test",
		Message:   "Test",
		Recipient: "user-delete",
		TenantID:  "tenant-delete",
	}
	ns.Send(notification)

	err := ns.Delete(notification.ID)
	if err != nil {
		t.Fatalf("Failed to delete notification: %v", err)
	}

	_, err = ns.Get(notification.ID)
	if err == nil {
		t.Error("Expected error for deleted notification")
	}
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	ns := service.NewNotificationService()

	ns.Send(&service.Notification{
		Type:      service.NotificationTypeAlert,
		Title:     "Unread 1",
		Message:   "Test",
		Recipient: "user-unread",
		TenantID:  "tenant-unread",
	})

	ns.Send(&service.Notification{
		Type:      service.NotificationTypeInfo,
		Title:     "Unread 2",
		Message:   "Test",
		Recipient: "user-unread",
		TenantID:  "tenant-unread",
	})

	count := ns.GetUnreadCount("user-unread")
	if count < 2 {
		t.Errorf("Expected at least 2 unread, got %d", count)
	}
}

func TestNotificationService_Broadcast(t *testing.T) {
	ns := service.NewNotificationService()

	err := ns.Broadcast("tenant-broadcast", "System Alert", "System maintenance scheduled")
	if err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	notifications := ns.ListByTenant("tenant-broadcast")
	if len(notifications) < 1 {
		t.Error("Expected broadcast notification")
	}
}

func TestNotificationService_SetChannel(t *testing.T) {
	ns := service.NewNotificationService()

	channel := &service.NotificationChannel{
		ID:       "channel-1",
		Name:     "Email",
		Type:     "email",
		TenantID: "tenant-channel",
		Config: map[string]interface{}{
			"smtp_host": "smtp.example.com",
		},
		Enabled: true,
	}

	err := ns.SetChannel(channel)
	if err != nil {
		t.Fatalf("Failed to set channel: %v", err)
	}
}

func TestNotificationService_GetChannels(t *testing.T) {
	ns := service.NewNotificationService()

	ns.SetChannel(&service.NotificationChannel{
		ID:       "ch-1",
		Name:     "Email",
		Type:     "email",
		TenantID: "tenant-channels",
		Enabled:  true,
	})

	ns.SetChannel(&service.NotificationChannel{
		ID:       "ch-2",
		Name:     "Slack",
		Type:     "slack",
		TenantID: "tenant-channels",
		Enabled:  true,
	})

	channels := ns.GetChannels("tenant-channels")
	if len(channels) < 2 {
		t.Errorf("Expected at least 2 channels, got %d", len(channels))
	}
}

func TestNotificationService_NotificationTypes(t *testing.T) {
	types := []service.NotificationType{
		service.NotificationTypeAlert,
		service.NotificationTypeInfo,
		service.NotificationTypeWarning,
		service.NotificationTypeError,
	}

	for _, nt := range types {
		if string(nt) == "" {
			t.Errorf("Notification type %s is empty", nt)
		}
	}
}

func TestNotificationService_NotificationToJSON(t *testing.T) {
	notification := &service.Notification{
		ID:        "notif-1",
		Type:      service.NotificationTypeAlert,
		Title:     "Test",
		Message:   "Test message",
		Recipient: "user-1",
		TenantID:  "tenant-1",
		Status:    service.NotificationStatusSent,
		CreatedAt: time.Now(),
	}

	data, err := notification.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}