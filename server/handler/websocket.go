package handler

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/accnotify/server/model"
	"github.com/accnotify/server/storage"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket client connection
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	deviceKey string
	deviceID  int64
}

// Hub manages all WebSocket connections
type Hub struct {
	clients    map[string]*Client // deviceKey -> Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	storage    *storage.SQLiteStorage
	mu         sync.RWMutex
}

// BroadcastMessage represents a message to be sent to a specific device
type BroadcastMessage struct {
	DeviceKey string
	Message   []byte
}

// NewHub creates a new WebSocket hub
func NewHub(storage *storage.SQLiteStorage) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
		storage:    storage,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Close existing connection if any
			if existing, ok := h.clients[client.deviceKey]; ok {
				close(existing.send)
				existing.conn.Close()
			}
			h.clients[client.deviceKey] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s", client.deviceKey)

			// Send undelivered messages
			go h.sendUndeliveredMessages(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if existing, ok := h.clients[client.deviceKey]; ok && existing == client {
				delete(h.clients, client.deviceKey)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s", client.deviceKey)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if client, ok := h.clients[msg.DeviceKey]; ok {
				select {
				case client.send <- msg.Message:
				default:
					// Client buffer full, disconnect
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, msg.DeviceKey)
					close(client.send)
					h.mu.Unlock()
					continue
				}
			}
			h.mu.RUnlock()
		}
	}
}

// sendUndeliveredMessages sends all undelivered messages to a newly connected client
func (h *Hub) sendUndeliveredMessages(client *Client) {
	messages, err := h.storage.GetUndeliveredMessages(client.deviceID)
	if err != nil {
		log.Printf("Error getting undelivered messages: %v", err)
		return
	}

	for _, msg := range messages {
		wsMsg := model.WSMessage{
			Type:      model.WSTypeMessage,
			ID:        msg.MessageID,
			Timestamp: msg.CreatedAt.Unix(),
			Data: map[string]interface{}{
				"title":             msg.Title,
				"body":              msg.Body,
				"group":             msg.Group,
				"icon":              msg.Icon,
				"url":               msg.URL,
				"sound":             msg.Sound,
				"badge":             msg.Badge,
				"encrypted_content": string(msg.EncryptedPayload),
			},
		}

		data, err := json.Marshal(wsMsg)
		if err != nil {
			continue
		}

		select {
		case client.send <- data:
		default:
			return
		}
	}
}

// SendToDevice sends a message to a specific device
func (h *Hub) SendToDevice(deviceKey string, msg *model.WSMessage) bool {
	data, err := json.Marshal(msg)
	if err != nil {
		return false
	}

	h.mu.RLock()
	_, online := h.clients[deviceKey]
	h.mu.RUnlock()

	if online {
		h.broadcast <- &BroadcastMessage{
			DeviceKey: deviceKey,
			Message:   data,
		}
	}

	return online
}

// IsOnline checks if a device is currently connected
func (h *Hub) IsOnline(deviceKey string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[deviceKey]
	return ok
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (ACKs, etc.)
		var wsMsg model.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case model.WSTypeAck:
			// Mark message as delivered
			if wsMsg.ID != "" {
				c.hub.storage.MarkMessageDelivered(wsMsg.ID)
			}
		case model.WSTypePong:
			// Client responded to ping
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			pingMsg := model.WSMessage{
				Type:      model.WSTypePing,
				Timestamp: time.Now().Unix(),
			}
			data, _ := json.Marshal(pingMsg)
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}
