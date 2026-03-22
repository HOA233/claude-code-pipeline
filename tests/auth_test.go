package tests

import (
	"testing"
	"time"

	"github.com/company/claude-pipeline/internal/service"
)

// Auth Service Tests

func TestAuthBasic_New(t *testing.T) {
	as := service.NewAuthService()
	if as == nil {
		t.Fatal("Expected non-nil auth service")
	}
}

func TestAuthService_Login(t *testing.T) {
	as := service.NewAuthService()

	// Create a user first
	user := &service.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		TenantID: "tenant-1",
	}
	as.CreateUser(user)

	session, err := as.Login("testuser", "hashedpassword", "tenant-1")
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	if session.Token == "" {
		t.Error("Expected token to be generated")
	}

	if session.UserID != "user-1" {
		t.Error("User ID mismatch")
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	as := service.NewAuthService()

	_, err := as.Login("nonexistent", "wrongpass", "tenant-1")
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
}

func TestAuthService_Logout(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "logout-user",
		Username: "logoutuser",
		Password: "pass",
		TenantID: "tenant-logout",
	}
	as.CreateUser(user)
	session, _ := as.Login("logoutuser", "pass", "tenant-logout")

	err := as.Logout(session.Token)
	if err != nil {
		t.Fatalf("Failed to logout: %v", err)
	}

	// Session should be invalid
	_, err = as.ValidateSession(session.Token)
	if err == nil {
		t.Error("Expected error for invalidated session")
	}
}

func TestAuthBasic_ValidateSession(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "validate-user",
		Username: "validateuser",
		Password: "pass",
		TenantID: "tenant-validate",
	}
	as.CreateUser(user)
	session, _ := as.Login("validateuser", "pass", "tenant-validate")

	validated, err := as.ValidateSession(session.Token)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}

	if validated.UserID != "validate-user" {
		t.Error("User ID mismatch")
	}
}

func TestAuthBasic_ValidateSession_Invalid(t *testing.T) {
	as := service.NewAuthService()

	_, err := as.ValidateSession("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestAuthService_CreateUser(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "create-user",
		Username: "createuser",
		Email:    "create@example.com",
		Password: "password",
		TenantID: "tenant-create",
	}

	err := as.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
}

func TestAuthService_CreateUser_Duplicate(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "dup-user",
		Username: "dupuser",
		Password: "pass",
		TenantID: "tenant-dup",
	}
	as.CreateUser(user)

	// Try to create with same username
	user2 := &service.User{
		ID:       "dup-user-2",
		Username: "dupuser",
		Password: "pass",
		TenantID: "tenant-dup",
	}

	err := as.CreateUser(user2)
	if err == nil {
		t.Error("Expected error for duplicate username")
	}
}

func TestAuthService_GetUser(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "get-user",
		Username: "getuser",
		Password: "pass",
		TenantID: "tenant-getuser",
	}
	as.CreateUser(user)

	retrieved, err := as.GetUser("get-user")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrieved.Username != "getuser" {
		t.Error("Username mismatch")
	}
}

func TestAuthService_UpdateUser(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "update-user",
		Username: "updateuser",
		Email:    "old@example.com",
		TenantID: "tenant-updateuser",
	}
	as.CreateUser(user)

	err := as.UpdateUser("update-user", map[string]interface{}{
		"email": "new@example.com",
	})

	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	retrieved, _ := as.GetUser("update-user")
	if retrieved.Email != "new@example.com" {
		t.Error("Email not updated")
	}
}

func TestAuthService_DeleteUser(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "delete-user",
		Username: "deleteuser",
		TenantID: "tenant-deleteuser",
	}
	as.CreateUser(user)

	err := as.DeleteUser("delete-user")
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, err = as.GetUser("delete-user")
	if err == nil {
		t.Error("Expected error for deleted user")
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "changepass-user",
		Username: "changepassuser",
		Password: "oldpass",
		TenantID: "tenant-changepass",
	}
	as.CreateUser(user)

	err := as.ChangePassword("changepass-user", "oldpass", "newpass")
	if err != nil {
		t.Fatalf("Failed to change password: %v", err)
	}
}

func TestAuthService_AssignRole(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "role-user",
		Username: "roleuser",
		TenantID: "tenant-role",
	}
	as.CreateUser(user)

	err := as.AssignRole("role-user", "admin")
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}
}

func TestAuthService_HasPermission(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "perm-user",
		Username: "permuser",
		TenantID: "tenant-perm",
		Roles:    []string{"admin"},
	}
	as.CreateUser(user)

	hasPerm := as.HasPermission("perm-user", "write")
	_ = hasPerm
}

func TestAuthService_ListUsers(t *testing.T) {
	as := service.NewAuthService()

	as.CreateUser(&service.User{
		ID:       "list-user-1",
		Username: "listuser1",
		TenantID: "tenant-listusers",
	})

	as.CreateUser(&service.User{
		ID:       "list-user-2",
		Username: "listuser2",
		TenantID: "tenant-listusers",
	})

	users := as.ListUsers("tenant-listusers")
	if len(users) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(users))
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	as := service.NewAuthService()

	user := &service.User{
		ID:       "refresh-user",
		Username: "refreshuser",
		Password: "pass",
		TenantID: "tenant-refresh",
	}
	as.CreateUser(user)
	session, _ := as.Login("refreshuser", "pass", "tenant-refresh")

	newSession, err := as.RefreshToken(session.Token)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if newSession.Token == session.Token {
		t.Error("New token should be different")
	}
}

func TestAuthService_UserToJSON(t *testing.T) {
	user := &service.User{
		ID:        "json-user",
		Username:  "jsonuser",
		Email:     "json@example.com",
		TenantID:  "tenant-1",
		CreatedAt: time.Now(),
	}

	data, err := user.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}