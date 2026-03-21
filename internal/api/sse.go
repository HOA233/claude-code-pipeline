package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SSEHandler handles Server-Sent Events connections
type SSEHandler struct {
	mu          sync.RWMutex
	clients     map[string]map[*sseClient]struct{}
	broadcast   chan *sseMessage
	register    chan *sseClient
	unregister  chan *sseClient
}

type sseClient struct {
	channel string
	send    chan []byte
}

type sseMessage struct {
	channel string
	event   string
	data    interface{}
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler() *SSEHandler {
	h := &SSEHandler{
		clients:    make(map[string]map[*sseClient]struct{}),
		broadcast:  make(chan *sseMessage, 1000),
		register:   make(chan *sseClient),
		unregister: make(chan *sseClient),
	}
	go h.run()
	return h
}

// run handles the main event loop for SSE
func (h *SSEHandler) run() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.channel]; !ok {
				h.clients[client.channel] = make(map[*sseClient]struct{})
			}
			h.clients[client.channel][client] = struct{}{}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.channel]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.clients, client.channel)
				}
			}
			h.mu.Unlock()
			close(client.send)

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg.data)
			if err != nil {
				continue
			}
			payload := fmt.Sprintf("event: %s\ndata: %s\n\n", msg.event, data)

			h.mu.RLock()
			clients, ok := h.clients[msg.channel]
			if ok {
				for client := range clients {
					select {
					case client.send <- []byte(payload):
					default:
						// Client buffer full, skip
					}
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			// Send keep-alive comments to all clients
			h.mu.RLock()
			for _, clients := range h.clients {
				for client := range clients {
					select {
					case client.send <- []byte(": keep-alive\n\n"):
					default:
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Handle handles SSE connections for a channel
func (h *SSEHandler) Handle(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel required"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	client := &sseClient{
		channel: channel,
		send:    make(chan []byte, 100),
	}

	// Register client
	h.register <- client
	defer func() {
		h.unregister <- client
	}()

	// Send initial connection message
	c.Writer.Write([]byte("event: connected\ndata: {\"status\":\"connected\"}\n\n"))
	c.Writer.Flush()

	// Stream events
	for {
		select {
		case data, ok := <-client.send:
			if !ok {
				return
			}
			c.Writer.Write(data)
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			return
		}
	}
}

// Publish sends an event to all clients subscribed to a channel
func (h *SSEHandler) Publish(channel, event string, data interface{}) {
	h.broadcast <- &sseMessage{
		channel: channel,
		event:   event,
		data:    data,
	}
}

// BroadcastTaskUpdate broadcasts a task update event
func (h *SSEHandler) BroadcastTaskUpdate(taskID string, status string, data interface{}) {
	h.Publish("task:"+taskID, "task.update", map[string]interface{}{
		"task_id": taskID,
		"status":  status,
		"data":    data,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// BroadcastPipelineUpdate broadcasts a pipeline update event
func (h *SSEHandler) BroadcastPipelineUpdate(pipelineID string, status string, data interface{}) {
	h.Publish("pipeline:"+pipelineID, "pipeline.update", map[string]interface{}{
		"pipeline_id": pipelineID,
		"status":      status,
		"data":        data,
		"time":        time.Now().UTC().Format(time.RFC3339),
	})
}

// BroadcastRunUpdate broadcasts a run update event
func (h *SSEHandler) BroadcastRunUpdate(runID string, step string, data interface{}) {
	h.Publish("run:"+runID, "run.step", map[string]interface{}{
		"run_id": runID,
		"step":   step,
		"data":   data,
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// BroadcastLog broadcasts a log message
func (h *SSEHandler) BroadcastLog(channel string, level string, message string) {
	h.Publish(channel, "log", map[string]interface{}{
		"level":   level,
		"message": message,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// GetClientCount returns the number of clients for a channel
func (h *SSEHandler) GetClientCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[channel]; ok {
		return len(clients)
	}
	return 0
}

// GetTotalClients returns the total number of connected clients
func (h *SSEHandler) GetTotalClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}

// GetStats returns SSE handler statistics
func (h *SSEHandler) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	channels := make(map[string]int)
	for channel, clients := range h.clients {
		channels[channel] = len(clients)
	}

	return map[string]interface{}{
		"total_clients": h.GetTotalClients(),
		"channels":      channels,
	}
}