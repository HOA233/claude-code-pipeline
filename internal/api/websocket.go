package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/company/claude-pipeline/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections for real-time updates
type WebSocketHandler struct {
	redis      *repository.RedisClient
	clients    map[string]map[*websocket.Conn]bool
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan *WebSocketMessage
	mu         sync.RWMutex
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	conn     *websocket.Conn
	taskID   string
	runID    string
	clientID string
}

// WebSocketMessage represents a message to broadcast
type WebSocketMessage struct {
	TaskID   string      `json:"task_id,omitempty"`
	RunID    string      `json:"run_id,omitempty"`
	Type     string      `json:"type"`
	Data     interface{} `json:"data"`
	Progress int         `json:"progress,omitempty"`
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(redis *repository.RedisClient) *WebSocketHandler {
	h := &WebSocketHandler{
		redis:      redis,
		clients:    make(map[string]map[*websocket.Conn]bool),
		register:   make(chan *WebSocketClient, 256),
		unregister: make(chan *WebSocketClient, 256),
		broadcast:  make(chan *WebSocketMessage, 1024),
	}

	go h.runHub()
	go h.subscribeToRedis()

	return h
}

// runHub manages WebSocket client connections
func (h *WebSocketHandler) runHub() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			key := h.getClientKey(client)
			if h.clients[key] == nil {
				h.clients[key] = make(map[*websocket.Conn]bool)
			}
			h.clients[key][client.conn] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			key := h.getClientKey(client)
			if h.clients[key] != nil {
				delete(h.clients[key], client.conn)
				if len(h.clients[key]) == 0 {
					delete(h.clients, key)
				}
			}
			h.mu.Unlock()
			client.conn.Close()

		case message := <-h.broadcast:
			h.mu.RLock()
			// Broadcast to task subscribers
			if message.TaskID != "" {
				key := "task:" + message.TaskID
				for conn := range h.clients[key] {
					if err := conn.WriteJSON(message); err != nil {
						conn.Close()
						delete(h.clients[key], conn)
					}
				}
			}
			// Broadcast to run subscribers
			if message.RunID != "" {
				key := "run:" + message.RunID
				for conn := range h.clients[key] {
					if err := conn.WriteJSON(message); err != nil {
						conn.Close()
						delete(h.clients[key], conn)
					}
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			// Send ping to all clients
			h.mu.RLock()
			for _, conns := range h.clients {
				for conn := range conns {
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						conn.Close()
						delete(conns, conn)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// subscribeToRedis subscribes to Redis pub/sub for task/run updates
func (h *WebSocketHandler) subscribeToRedis() {
	// This would subscribe to Redis channels for task/run updates
	// For now, this is a placeholder
}

// getClientKey generates a key for client grouping
func (h *WebSocketHandler) getClientKey(client *WebSocketClient) string {
	if client.taskID != "" {
		return "task:" + client.taskID
	}
	if client.runID != "" {
		return "run:" + client.runID
	}
	return "global"
}

// HandleTaskWS handles WebSocket connection for task updates
func (h *WebSocketHandler) HandleTaskWS(c *gin.Context) {
	taskID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &WebSocketClient{
		conn:   conn,
		taskID: taskID,
	}

	h.register <- client

	// Send initial status
	task, err := h.redis.GetTask(c.Request.Context(), taskID)
	if err == nil {
		conn.WriteJSON(&WebSocketMessage{
			TaskID: taskID,
			Type:   "task:status",
			Data:   task,
		})
	}

	// Read messages from client (for pong, etc.)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			h.unregister <- client
			break
		}
	}
}

// HandleRunWS handles WebSocket connection for run updates
func (h *WebSocketHandler) HandleRunWS(c *gin.Context) {
	runID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &WebSocketClient{
		conn:  conn,
		runID: runID,
	}

	h.register <- client

	// Send initial status
	run, err := h.redis.GetRun(c.Request.Context(), runID)
	if err == nil {
		conn.WriteJSON(&WebSocketMessage{
			RunID: runID,
			Type:  "run:status",
			Data:  run,
		})
	}

	// Read messages from client
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			h.unregister <- client
			break
		}
	}
}

// HandleGlobalWS handles WebSocket connection for global updates
func (h *WebSocketHandler) HandleGlobalWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &WebSocketClient{
		conn: conn,
	}

	h.register <- client

	// Read messages from client
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			h.unregister <- client
			break
		}
	}
}

// BroadcastTaskUpdate broadcasts a task update to all subscribers
func (h *WebSocketHandler) BroadcastTaskUpdate(taskID string, updateType string, data interface{}) {
	h.broadcast <- &WebSocketMessage{
		TaskID: taskID,
		Type:   updateType,
		Data:   data,
	}
}

// BroadcastRunUpdate broadcasts a run update to all subscribers
func (h *WebSocketHandler) BroadcastRunUpdate(runID string, updateType string, data interface{}) {
	h.broadcast <- &WebSocketMessage{
		RunID: runID,
		Type:  updateType,
		Data:  data,
	}
}

// BroadcastTaskOutput broadcasts task output chunk
func (h *WebSocketHandler) BroadcastTaskOutput(taskID string, output string, progress int) {
	h.broadcast <- &WebSocketMessage{
		TaskID:   taskID,
		Type:     "task:output",
		Data:     output,
		Progress: progress,
	}
}

// BroadcastPipelineUpdate broadcasts a pipeline update
func (h *WebSocketHandler) BroadcastPipelineUpdate(pipelineID string, updateType string, data interface{}) {
	h.broadcast <- &WebSocketMessage{
		Type: updateType,
		Data: gin.H{
			"pipeline_id": pipelineID,
			"details":     data,
		},
	}
}

// ConnectionStats returns current connection statistics
func (h *WebSocketHandler) ConnectionStats() map[string]int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]int)
	for key, conns := range h.clients {
		stats[key] = len(conns)
	}
	return stats
}

// TaskSSEHandler handles Server-Sent Events for clients that prefer SSE over WebSocket
type TaskSSEHandler struct {
	redis *repository.RedisClient
}

// NewTaskSSEHandler creates a new task SSE handler
func NewTaskSSEHandler(redis *repository.RedisClient) *TaskSSEHandler {
	return &TaskSSEHandler{
		redis: redis,
	}
}

// HandleTaskSSE handles SSE connection for task updates
func (h *TaskSSEHandler) HandleTaskSSE(c *gin.Context) {
	taskID := c.Param("id")

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Send initial status
	task, err := h.redis.GetTask(c.Request.Context(), taskID)
	if err == nil {
		data, _ := json.Marshal(map[string]interface{}{
			"type": "task:status",
			"data": task,
		})
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		c.Writer.Flush()
	}

	// Subscribe to task updates
	ctx := c.Request.Context()
	pubsub := h.redis.SubscribeTaskUpdates(ctx, taskID)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", msg.Payload)
			c.Writer.Flush()
		case <-time.After(30 * time.Second):
			// Send keepalive
			fmt.Fprintf(c.Writer, ": keepalive\n\n")
			c.Writer.Flush()
		}
	}
}

// HandleRunSSE handles SSE connection for run updates
func (h *TaskSSEHandler) HandleRunSSE(c *gin.Context) {
	runID := c.Param("id")

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Send initial status
	run, err := h.redis.GetRun(c.Request.Context(), runID)
	if err == nil {
		data, _ := json.Marshal(map[string]interface{}{
			"type": "run:status",
			"data": run,
		})
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		c.Writer.Flush()
	}

	// Subscribe to run updates
	ctx := c.Request.Context()
	pubsub := h.redis.SubscribeRunUpdates(ctx, runID)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", msg.Payload)
			c.Writer.Flush()
		case <-time.After(30 * time.Second):
			// Send keepalive
			fmt.Fprintf(c.Writer, ": keepalive\n\n")
			c.Writer.Flush()
		}
	}
}

// HandleGlobalSSE handles SSE connection for global updates
func (h *TaskSSEHandler) HandleGlobalSSE(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	ctx := c.Request.Context()

	// Send periodic system stats
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := h.getSystemStats()
			data, _ := json.Marshal(map[string]interface{}{
				"type": "system:stats",
				"data": stats,
			})
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		}
	}
}

func (h *TaskSSEHandler) getSystemStats() map[string]interface{} {
	ctx := context.Background()
	stats := make(map[string]interface{})

	// Get queue lengths
	taskQueueLen, _ := h.redis.GetQueueLength(ctx)
	runQueueLen, _ := h.redis.GetRunQueueLength(ctx)

	stats["task_queue_length"] = taskQueueLen
	stats["run_queue_length"] = runQueueLen

	return stats
}