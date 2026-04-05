package handler

import (
	"encoding/json"
<<<<<<< HEAD
	"log"
	"net/http"
	"strconv"
	"strings"
=======
	"net/http"
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
<<<<<<< HEAD
	"github.com/accnotify/server/apns"
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
	"github.com/accnotify/server/crypto"
	"github.com/accnotify/server/model"
	"github.com/accnotify/server/storage"
)

// PushHandler handles push notification requests
type PushHandler struct {
	storage *storage.SQLiteStorage
	hub     *Hub
	crypto  *crypto.Crypto
<<<<<<< HEAD
	apns    *apns.Service
}

// NewPushHandler creates a new push handler
func NewPushHandler(storage *storage.SQLiteStorage, hub *Hub, apnsService *apns.Service) *PushHandler {
=======
}

// NewPushHandler creates a new push handler
func NewPushHandler(storage *storage.SQLiteStorage, hub *Hub) *PushHandler {
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
	return &PushHandler{
		storage: storage,
		hub:     hub,
		crypto:  crypto.NewCrypto(),
<<<<<<< HEAD
		apns:    apnsService,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
	}
}

// HandlePush handles POST /push/:device_key
func (h *PushHandler) HandlePush(c *gin.Context) {
	deviceKey := c.Param("device_key")
	if deviceKey == "" {
		c.JSON(http.StatusBadRequest, model.PushResponse{
			Success: false,
			Error:   "Missing device key",
		})
		return
	}

	// Get device
	device, err := h.storage.GetDeviceByKey(deviceKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.PushResponse{
<<<<<<< HEAD
			Code:    500,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
			Success: false,
			Error:   "Database error",
		})
		return
	}
	if device == nil {
		c.JSON(http.StatusNotFound, model.PushResponse{
<<<<<<< HEAD
			Code:    404,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
			Success: false,
			Error:   "Device not found",
		})
		return
	}

	// Parse request
	var req model.PushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.PushResponse{
<<<<<<< HEAD
			Code:    400,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Generate message ID
	messageID := uuid.New().String()

	// Create message
	msg := &model.Message{
		DeviceID:  device.ID,
		MessageID: messageID,
		Title:     req.Title,
		Body:      req.Body,
		Group:     req.Group,
		Icon:      req.Icon,
		URL:       req.URL,
		Sound:     req.Sound,
		Badge:     req.Badge,
	}

	// Encrypt if device has public key (E2E mode)
	var encryptedContent string
	if device.PublicKey != "" {
		publicKey, err := h.crypto.ParsePublicKey(device.PublicKey)
		if err == nil {
			// Create payload to encrypt
			payload := map[string]interface{}{
				"title": req.Title,
				"body":  req.Body,
				"group": req.Group,
				"icon":  req.Icon,
				"url":   req.URL,
				"sound": req.Sound,
				"badge": req.Badge,
			}
			payloadBytes, _ := json.Marshal(payload)
			encryptedContent, _ = h.crypto.EncryptMessage(publicKey, payloadBytes)
			msg.EncryptedPayload = []byte(encryptedContent)
		}
	}

	// Store message
	if err := h.storage.CreateMessage(msg); err != nil {
		c.JSON(http.StatusInternalServerError, model.PushResponse{
<<<<<<< HEAD
			Code:    500,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
			Success: false,
			Error:   "Failed to store message",
		})
		return
	}

	// Prepare WebSocket message
	wsMsg := &model.WSMessage{
		Type:      model.WSTypeMessage,
		ID:        messageID,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"title":             req.Title,
			"body":              req.Body,
			"group":             req.Group,
			"icon":              req.Icon,
			"url":               req.URL,
			"sound":             req.Sound,
			"badge":             req.Badge,
			"encrypted_content": encryptedContent,
		},
	}

	// Send via WebSocket
	delivered := h.hub.SendToDevice(deviceKey, wsMsg)

	if delivered {
		h.storage.MarkMessageDelivered(messageID)
	}

<<<<<<< HEAD
	// Send via APNs if device has token and APNs is enabled
	if h.apns != nil && h.apns.IsEnabled() && device.DeviceToken != "" && device.Platform == "ios" {
		log.Printf("[APNs] Attempting to send push to device %s (Token: %s...)", deviceKey, device.DeviceToken[:10])
		
		// Map Bark level to APNs options
		level := c.Query("level")
		if level == "" {
			level = req.Group // Some users might pass level in other fields
		}

		apnsOptions := map[string]interface{}{
			"sound":           req.Sound,
			"badge":           req.Badge,
			"mutable-content": true, // Critical for Bark history and custom icons
		}

		// Handle sound specifically for Bark
		if req.Sound != "" {
			if !strings.HasSuffix(req.Sound, ".caf") {
				apnsOptions["sound"] = req.Sound + ".caf"
			}
		} else {
			apnsOptions["sound"] = "1" // Default Bark sound
		}
		
		if req.Group != "" {
			apnsOptions["thread-id"] = req.Group
			apnsOptions["group"] = req.Group
		}
		
		if req.URL != "" {
			apnsOptions["url"] = req.URL
		}
		
		if req.Icon != "" {
			apnsOptions["icon"] = req.Icon
		}

		// Handle notification level (active, timeSensitive, passive, critical)
		if level != "" {
			apnsOptions["level"] = level
		}
		
		if err := h.apns.Push(device.DeviceToken, req.Title, req.Body, apnsOptions); err != nil {
			log.Printf("[APNs] Failed to send push to %s: %v", deviceKey, err)
		} else {
			log.Printf("[APNs] Push sent successfully to %s", deviceKey)
			if !delivered {
				h.storage.MarkMessageDelivered(messageID)
				delivered = true
			}
		}
	} else {
		log.Printf("[APNs] Skip sending: apns_enabled=%v, has_token=%v, platform=%s", 
			h.apns != nil && h.apns.IsEnabled(), device.DeviceToken != "", device.Platform)
	}

	c.JSON(http.StatusOK, model.PushResponse{
		Code:      200,
=======
	c.JSON(http.StatusOK, model.PushResponse{
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
		Success:   true,
		MessageID: messageID,
	})
}

// HandleSimplePush handles GET /push/:device_key/:title/:body (Bark-compatible)
func (h *PushHandler) HandleSimplePush(c *gin.Context) {
	deviceKey := c.Param("device_key")
	title := c.Param("title")
	body := c.Param("body")

	// If only one param, treat it as body
	if body == "" {
		body = title
		title = "Accnotify"
	}

<<<<<<< HEAD
	// Get other parameters from query string (Bark compatibility)
	sound := c.Query("sound")
	group := c.Query("group")
	if group == "" {
		group = c.Query("thread-id")
	}
	icon := c.Query("icon")
	url := c.Query("url")
	badgeStr := c.Query("badge")
	badge := 0
	if badgeStr != "" {
		if b, err := strconv.Atoi(badgeStr); err == nil {
			badge = b
		}
	}

=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
	// Convert to POST request
	c.Set("device_key", deviceKey)

	// Get device
	device, err := h.storage.GetDeviceByKey(deviceKey)
	if err != nil || device == nil {
		c.JSON(http.StatusNotFound, model.PushResponse{
<<<<<<< HEAD
			Code:    404,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
			Success: false,
			Error:   "Device not found",
		})
		return
	}

	messageID := uuid.New().String()

	msg := &model.Message{
		DeviceID:  device.ID,
		MessageID: messageID,
		Title:     title,
		Body:      body,
<<<<<<< HEAD
		Group:     group,
		Icon:      icon,
		URL:       url,
		Sound:     sound,
		Badge:     badge,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
	}

	// Encrypt if device has public key
	var encryptedContent string
	if device.PublicKey != "" {
		publicKey, err := h.crypto.ParsePublicKey(device.PublicKey)
		if err == nil {
			payload := map[string]interface{}{
				"title": title,
				"body":  body,
<<<<<<< HEAD
				"group": group,
				"icon":  icon,
				"url":   url,
				"sound": sound,
				"badge": badge,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
			}
			payloadBytes, _ := json.Marshal(payload)
			encryptedContent, _ = h.crypto.EncryptMessage(publicKey, payloadBytes)
			msg.EncryptedPayload = []byte(encryptedContent)
		}
	}

	if err := h.storage.CreateMessage(msg); err != nil {
		c.JSON(http.StatusInternalServerError, model.PushResponse{
<<<<<<< HEAD
			Code:    500,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
			Success: false,
			Error:   "Failed to store message",
		})
		return
	}

	wsMsg := &model.WSMessage{
		Type:      model.WSTypeMessage,
		ID:        messageID,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"title":             title,
			"body":              body,
<<<<<<< HEAD
			"group":             group,
			"icon":              icon,
			"url":               url,
			"sound":             sound,
			"badge":             badge,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
			"encrypted_content": encryptedContent,
		},
	}

	delivered := h.hub.SendToDevice(deviceKey, wsMsg)
	if delivered {
		h.storage.MarkMessageDelivered(messageID)
	}

<<<<<<< HEAD
	// Send via APNs if device has token and APNs is enabled
	if h.apns != nil && h.apns.IsEnabled() && device.DeviceToken != "" && device.Platform == "ios" {
		log.Printf("[APNs] Attempting to send simple push to device %s (Token: %s...)", deviceKey, device.DeviceToken[:10])
		
		// Map Bark level to APNs options
		level := c.Query("level")

		apnsOptions := map[string]interface{}{
			"sound":           sound,
			"badge":           badge,
			"mutable-content": true, // Critical for Bark history and custom icons
		}

		// Handle sound specifically for Bark
		if sound != "" {
			if !strings.HasSuffix(sound, ".caf") {
				apnsOptions["sound"] = sound + ".caf"
			}
		} else {
			apnsOptions["sound"] = "1" // Default Bark sound
		}
		
		if group != "" {
			apnsOptions["thread-id"] = group
			apnsOptions["group"] = group
		}
		
		if url != "" {
			apnsOptions["url"] = url
		}
		
		if icon != "" {
			apnsOptions["icon"] = icon
		}

		// Handle notification level (active, timeSensitive, passive, critical)
		if level != "" {
			apnsOptions["level"] = level
		}

		if err := h.apns.Push(device.DeviceToken, title, body, apnsOptions); err != nil {
			log.Printf("[APNs] Failed to send push to %s: %v", deviceKey, err)
		} else {
			log.Printf("[APNs] Push sent successfully to %s", deviceKey)
			if !delivered {
				h.storage.MarkMessageDelivered(messageID)
				delivered = true
			}
		}
	} else {
		log.Printf("[APNs] Skip sending: apns_enabled=%v, has_token=%v, platform=%s", 
			h.apns != nil && h.apns.IsEnabled(), device.DeviceToken != "", device.Platform)
	}

	c.JSON(http.StatusOK, model.PushResponse{
		Code:      200,
=======
	c.JSON(http.StatusOK, model.PushResponse{
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
		Success:   true,
		MessageID: messageID,
	})
}

<<<<<<< HEAD
// HandleRegister handles POST/GET /register
func (h *PushHandler) HandleRegister(c *gin.Context) {
	var req model.RegisterRequest
	
	// Support both JSON body and query parameters
	if c.Request.Method == "POST" {
		if err := c.ShouldBindJSON(&req); err != nil {
			// Try query parameters if JSON binding fails
			req.DeviceKey = c.Query("device_key")
			if req.DeviceKey == "" {
				req.DeviceKey = c.Query("key")
			}
			req.PublicKey = c.Query("public_key")
			req.Name = c.Query("name")
			req.DeviceToken = c.Query("devicetoken")
			req.Platform = c.Query("platform")
		}
	} else {
		// For GET requests, use query parameters
		req.DeviceKey = c.Query("device_key")
		if req.DeviceKey == "" {
			req.DeviceKey = c.Query("key")
		}
		req.PublicKey = c.Query("public_key")
		req.Name = c.Query("name")
		req.DeviceToken = c.Query("devicetoken")
		req.Platform = c.Query("platform")
	}

	// For Bark compatibility, if device_key is missing, generate one
	if req.DeviceKey == "" && req.DeviceToken != "" {
		req.DeviceKey = uuid.New().String()
	}

	if req.DeviceKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing device key",
=======
// HandleRegister handles POST /register
func (h *PushHandler) HandleRegister(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
		})
		return
	}

<<<<<<< HEAD
	// For Bark compatibility, if platform is not specified but devicetoken is present, assume ios
	if req.Platform == "" && req.DeviceToken != "" {
		req.Platform = "ios"
	}

=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
	// Check if device exists
	device, err := h.storage.GetDeviceByKey(req.DeviceKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	if device != nil {
<<<<<<< HEAD
		// Update device information
		updates := false
		
		// Update public key if provided
=======
		// Update public key
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
		if req.PublicKey != "" {
			if err := h.storage.UpdateDevicePublicKey(req.DeviceKey, req.PublicKey); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to update public key",
				})
				return
			}
<<<<<<< HEAD
			updates = true
		}
		
		// Update device token if provided (for APNs)
		if req.DeviceToken != "" {
			if err := h.storage.UpdateDeviceToken(req.DeviceKey, req.DeviceToken); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to update device token",
				})
				return
			}
			updates = true
		}
		
		// Update platform if provided
		if req.Platform != "" {
			if err := h.storage.UpdateDevicePlatform(req.DeviceKey, req.Platform); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to update platform",
				})
				return
			}
			updates = true
		}
		
		// Update name if provided
		if req.Name != "" {
			if err := h.storage.UpdateDeviceName(req.DeviceKey, req.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to update name",
				})
				return
			}
			updates = true
		}
		
		message := "Device updated"
		if !updates {
			message = "Device already registered"
		}
		
		// Bark-compatible response format
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": message,
			"data": gin.H{
				"key":        req.DeviceKey,
				"device_key": req.DeviceKey,
			},
=======
		}
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"device_key": req.DeviceKey,
			"message":    "Device updated",
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
		})
		return
	}

	// Create new device
	newDevice := &model.Device{
<<<<<<< HEAD
		DeviceKey:   req.DeviceKey,
		PublicKey:   req.PublicKey,
		Name:        req.Name,
		DeviceToken: req.DeviceToken,
		Platform:    req.Platform,
=======
		DeviceKey: req.DeviceKey,
		PublicKey: req.PublicKey,
		Name:      req.Name,
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
	}

	if err := h.storage.CreateDevice(newDevice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create device",
		})
		return
	}

<<<<<<< HEAD
	// Bark-compatible response format
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Device registered",
		"data": gin.H{
			"key":        req.DeviceKey,
			"device_key": req.DeviceKey,
		},
	})
}

// HandleHealth handles GET /health, /ping, /
func (h *PushHandler) HandleHealth(c *gin.Context) {
	log.Printf("[HealthCheck] Ping received from %s", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "pong", // Bark compatibility for /ping
		"status":  "ok",
		"data": gin.H{
			"version": "1.0.0",
		},
=======
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"device_key": req.DeviceKey,
		"message":    "Device registered",
	})
}

// HandleHealth handles GET /health
func (h *PushHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
		"timestamp": time.Now().Unix(),
	})
}
