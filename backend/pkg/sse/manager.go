package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

// Event represents a server-sent event
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client represents a connected client
type Client struct {
	UserID string
	Send   chan []byte
}

// Manager manages SSE connections
type Manager struct {
	clients    map[*Client]bool
	userClients map[string][]*Client // Map userID to list of clients (multiple tabs/devices)
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mutex      sync.RWMutex
}

type BroadcastMessage struct {
	UserID  string
	Message []byte
}

// NewManager creates a new SSE manager
func NewManager() *Manager {
	return &Manager{
		clients:     make(map[*Client]bool),
		userClients: make(map[string][]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage),
	}
}

// Run starts the SSE manager loop
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mutex.Lock()
			m.clients[client] = true
			m.userClients[client.UserID] = append(m.userClients[client.UserID], client)
			m.mutex.Unlock()
			log.Printf("Client connected: %s", client.UserID)

		case client := <-m.unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.Send)
				
				// Remove from userClients
				clients := m.userClients[client.UserID]
				for i, c := range clients {
					if c == client {
						m.userClients[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(m.userClients[client.UserID]) == 0 {
					delete(m.userClients, client.UserID)
				}
			}
			m.mutex.Unlock()
			log.Printf("Client disconnected: %s", client.UserID)

		case message := <-m.broadcast:
			m.mutex.RLock()
			clients, ok := m.userClients[message.UserID]
			m.mutex.RUnlock()
			
			if ok {
				for _, client := range clients {
					select {
					case client.Send <- message.Message:
					default:
						close(client.Send)
						delete(m.clients, client)
					}
				}
			}
		}
	}
}

// ServeHTTP handles the SSE endpoint
func (m *Manager) ServeHTTP(c *gin.Context, userID string) {
	client := &Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
	}

	m.register <- client

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Send initial connection message
	c.SSEvent("connected", map[string]string{
		"message": "Connected to notification service",
	})
	c.Writer.Flush()

	defer func() {
		m.unregister <- client
	}()

	notify := c.Writer.CloseNotify()
	for {
		select {
		case <-notify:
			return
		case message, ok := <-client.Send:
			if !ok {
				return
			}
			c.Writer.Write(message)
			c.Writer.Flush()
		}
	}
}

// SendToUser sends a message to a specific user
func (m *Manager) SendToUser(userID string, eventType string, payload interface{}) {
	data, err := json.Marshal(Event{
		Type:    eventType,
		Payload: payload,
	})
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	// Format as SSE message: "data: ...\n\n"
	message := []byte(fmt.Sprintf("data: %s\n\n", data))

	m.broadcast <- &BroadcastMessage{
		UserID:  userID,
		Message: message,
	}
}
