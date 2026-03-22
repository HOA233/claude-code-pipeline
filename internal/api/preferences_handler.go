package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// UserPreferences 用户偏好设置
type UserPreferences struct {
	Theme           string                 `json:"theme"`
	Language        string                 `json:"language"`
	SidebarCollapsed bool                  `json:"sidebar_collapsed"`
	AutoRefresh     bool                   `json:"auto_refresh"`
	RefreshInterval int                    `json:"refresh_interval"`
	Notifications   NotificationPreferences `json:"notifications"`
	Dashboard       DashboardPreferences   `json:"dashboard"`
	Shortcuts       map[string]string      `json:"shortcuts"`
}

// NotificationPreferences 通知偏好
type NotificationPreferences struct {
	Enabled       bool     `json:"enabled"`
	Email         bool     `json:"email"`
	Desktop       bool     `json:"desktop"`
	ExecutionDone bool     `json:"execution_done"`
	Errors        bool     `json:"errors"`
	ScheduledJobs bool     `json:"scheduled_jobs"`
	Sound         bool     `json:"sound"`
	ExcludeTypes  []string `json:"exclude_types"`
}

// DashboardPreferences 仪表盘偏好
type DashboardPreferences struct {
	DefaultView    string   `json:"default_view"`
	HiddenCards    []string `json:"hidden_cards"`
	CardOrder      []string `json:"card_order"`
	ChartsEnabled  bool     `json:"charts_enabled"`
	TimelineLimit  int      `json:"timeline_limit"`
	ShowQuickStats bool     `json:"show_quick_stats"`
}

// UserPreferencesHandler 用户偏好处理器
type UserPreferencesHandler struct {
	prefs map[string]UserPreferences
	mu    sync.RWMutex
}

// NewUserPreferencesHandler 创建用户偏好处理器
func NewUserPreferencesHandler() *UserPreferencesHandler {
	return &UserPreferencesHandler{
		prefs: map[string]UserPreferences{
			"default": {
				Theme:           "dark",
				Language:        "zh-CN",
				SidebarCollapsed: false,
				AutoRefresh:     true,
				RefreshInterval: 10,
				Notifications: NotificationPreferences{
					Enabled:       true,
					Email:         false,
					Desktop:       true,
					ExecutionDone: true,
					Errors:        true,
					ScheduledJobs: true,
					Sound:         false,
					ExcludeTypes:  []string{},
				},
				Dashboard: DashboardPreferences{
					DefaultView:    "overview",
					HiddenCards:    []string{},
					CardOrder:      []string{"stats", "charts", "health", "recent", "quick-actions", "timeline"},
					ChartsEnabled:  true,
					TimelineLimit:  5,
					ShowQuickStats: true,
				},
				Shortcuts: map[string]string{
					"navigate_home":      "g h",
					"navigate_agents":    "g a",
					"navigate_workflows": "g w",
					"refresh":            "r",
					"search":             "/",
					"command_palette":    "ctrl+k",
				},
			},
		},
	}
}

// GetPreferences 获取用户偏好
func (h *UserPreferencesHandler) GetPreferences(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	h.mu.RLock()
	defer h.mu.RUnlock()

	prefs, exists := h.prefs[userID]
	if !exists {
		prefs = h.prefs["default"]
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdatePreferences 更新用户偏好
func (h *UserPreferencesHandler) UpdatePreferences(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	var updates UserPreferences
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	h.prefs[userID] = updates
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message": "preferences updated",
		"prefs":   updates,
	})
}

// UpdateTheme 更新主题
func (h *UserPreferencesHandler) UpdateTheme(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	var req struct {
		Theme string `json:"theme"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	prefs, exists := h.prefs[userID]
	if !exists {
		prefs = h.prefs["default"]
	}
	prefs.Theme = req.Theme
	h.prefs[userID] = prefs

	c.JSON(http.StatusOK, gin.H{
		"message": "theme updated",
		"theme":   req.Theme,
	})
}

// UpdateNotifications 更新通知偏好
func (h *UserPreferencesHandler) UpdateNotifications(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	var req NotificationPreferences
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	prefs, exists := h.prefs[userID]
	if !exists {
		prefs = h.prefs["default"]
	}
	prefs.Notifications = req
	h.prefs[userID] = prefs

	c.JSON(http.StatusOK, gin.H{
		"message":      "notification preferences updated",
		"notifications": req,
	})
}

// UpdateDashboard 更新仪表盘偏好
func (h *UserPreferencesHandler) UpdateDashboard(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	var req DashboardPreferences
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	prefs, exists := h.prefs[userID]
	if !exists {
		prefs = h.prefs["default"]
	}
	prefs.Dashboard = req
	h.prefs[userID] = prefs

	c.JSON(http.StatusOK, gin.H{
		"message":   "dashboard preferences updated",
		"dashboard": req,
	})
}

// ExportPreferences 导出偏好设置
func (h *UserPreferencesHandler) ExportPreferences(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	h.mu.RLock()
	defer h.mu.RUnlock()

	prefs, exists := h.prefs[userID]
	if !exists {
		prefs = h.prefs["default"]
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=user-preferences.json")
	c.JSON(http.StatusOK, gin.H{
		"preferences": prefs,
		"exported_at": "now",
		"version":     "1.0",
	})
}

// ImportPreferences 导入偏好设置
func (h *UserPreferencesHandler) ImportPreferences(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	var req struct {
		Preferences UserPreferences `json:"preferences"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	h.prefs[userID] = req.Preferences
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message":     "preferences imported",
		"preferences": req.Preferences,
	})
}

// ResetPreferences 重置偏好设置
func (h *UserPreferencesHandler) ResetPreferences(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.prefs, userID)

	c.JSON(http.StatusOK, gin.H{
		"message": "preferences reset to default",
	})
}

// GetShortcuts 获取快捷键设置
func (h *UserPreferencesHandler) GetShortcuts(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	h.mu.RLock()
	defer h.mu.RUnlock()

	prefs, exists := h.prefs[userID]
	if !exists {
		prefs = h.prefs["default"]
	}

	c.JSON(http.StatusOK, gin.H{
		"shortcuts": prefs.Shortcuts,
	})
}

// UpdateShortcuts 更新快捷键设置
func (h *UserPreferencesHandler) UpdateShortcuts(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	prefs, exists := h.prefs[userID]
	if !exists {
		prefs = h.prefs["default"]
	}
	prefs.Shortcuts = req
	h.prefs[userID] = prefs

	c.JSON(http.StatusOK, gin.H{
		"message":   "shortcuts updated",
		"shortcuts": req,
	})
}

// ToJSON converts preferences to JSON
func (p *UserPreferences) ToJSON() string {
	data, _ := json.Marshal(p)
	return string(data)
}