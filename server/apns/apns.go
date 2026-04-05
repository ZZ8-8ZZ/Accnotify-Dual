package apns

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

// Service handles APNs push notifications
type Service struct {
	client      *apns2.Client
	enabled     bool
	topic       string
	development bool
}

// Config holds APNs configuration
type Config struct {
	Enabled     bool
	CertFile    string
	KeyFile     string
	KeyID       string
	TeamID      string
	Topic       string
	Development bool
}

// NewService creates a new APNs service
func NewService(cfg *Config) (*Service, error) {
	if !cfg.Enabled {
		return &Service{
			enabled: false,
		}, nil
	}

	var client *apns2.Client

	// Try token-based authentication first (recommended by Apple)
	if cfg.KeyID != "" && cfg.TeamID != "" {
		authKey, err := token.AuthKeyFromFile(cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load auth key: %w", err)
		}

		t := &token.Token{
			AuthKey: authKey,
			KeyID:   cfg.KeyID,
			TeamID:  cfg.TeamID,
		}

		if cfg.Development {
			client = apns2.NewTokenClient(t).Development()
		} else {
			client = apns2.NewTokenClient(t).Production()
		}
	} else if cfg.CertFile != "" {
		// Fall back to certificate-based authentication
		cert, err := certificate.FromP12File(cfg.CertFile, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate: %w", err)
		}

		if cfg.Development {
			client = apns2.NewClient(cert).Development()
		} else {
			client = apns2.NewClient(cert).Production()
		}
	} else {
		return nil, fmt.Errorf("no valid APNs credentials provided")
	}

	return &Service{
		client:      client,
		enabled:     true,
		topic:       cfg.Topic,
		development: cfg.Development,
	}, nil
}

// Push sends a push notification to a device
func (s *Service) Push(deviceToken string, title, body string, options map[string]interface{}) error {
	if !s.enabled {
		return fmt.Errorf("APNs service is not enabled")
	}

	if s.topic == "" {
		return fmt.Errorf("APNs topic is not configured")
	}

	// Create payload
	p := payload.NewPayload().
		AlertTitle(title).
		AlertBody(body)

	// Add optional fields
	if sound, ok := options["sound"].(string); ok && sound != "" {
		p.Sound(sound)
	}

	if badge, ok := options["badge"].(int); ok && badge > 0 {
		p.Badge(badge)
	}

	if category, ok := options["category"].(string); ok && category != "" {
		p.Category(category)
	}

	if threadID, ok := options["thread-id"].(string); ok && threadID != "" {
		p.ThreadID(threadID)
	}

	// Add custom fields to payload
	if customData, ok := options["custom"].(map[string]interface{}); ok {
		for key, value := range customData {
			p.Custom(key, value)
		}
	}

	// Add mutable-content if needed
	if mutableContent, ok := options["mutable-content"].(bool); ok && mutableContent {
		p.MutableContent()
	}

	// Add content-available for silent notifications
	if contentAvailable, ok := options["content-available"].(bool); ok && contentAvailable {
		p.ContentAvailable()
	}

	// Add URL if provided
	if url, ok := options["url"].(string); ok && url != "" {
		p.Custom("url", url)
	}

	// Add group if provided
	if group, ok := options["group"].(string); ok && group != "" {
		p.Custom("group", group)
	}

	// Add icon if provided
	if icon, ok := options["icon"].(string); ok && icon != "" {
		p.Custom("icon", icon)
	}

	// Add level if provided (Bark compatibility)
	if level, ok := options["level"].(string); ok && level != "" {
		p.Custom("level", level)
		if level == "critical" {
			p.CriticalAlert(sound, 1.0)
		}
	}

	// Create notification
	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       s.topic,
		Payload:     p,
	}

	// Send notification
	resp, err := s.client.Push(notification)
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("APNs error: %s (status: %d)", resp.Reason, resp.StatusCode)
	}

	log.Printf("APNs push sent successfully to device: %s, ID: %s", deviceToken, resp.ApnsID)
	return nil
}

// PushWithCustomPayload sends a push notification with custom payload
func (s *Service) PushWithCustomPayload(deviceToken string, customPayload interface{}) error {
	if !s.enabled {
		return fmt.Errorf("APNs service is not enabled")
	}

	if s.topic == "" {
		return fmt.Errorf("APNs topic is not configured")
	}

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       s.topic,
		Payload:     customPayload,
	}

	resp, err := s.client.Push(notification)
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("APNs error: %s (status: %d)", resp.Reason, resp.StatusCode)
	}

	log.Printf("APNs custom push sent successfully to device: %s, ID: %s", deviceToken, resp.ApnsID)
	return nil
}

// IsEnabled returns whether the APNs service is enabled
func (s *Service) IsEnabled() bool {
	return s.enabled
}

// GetTopic returns the configured APNs topic
func (s *Service) GetTopic() string {
	return s.topic
}

// IsDevelopment returns whether this is development mode
func (s *Service) IsDevelopment() bool {
	return s.development
}

// Close closes the APNs service
func (s *Service) Close() error {
	if s.client != nil && s.client.HTTPClient != nil {
		if transport, ok := s.client.HTTPClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	return nil
}

// ValidateConfig validates the APNs configuration
func ValidateConfig(cfg *Config) error {
	if !cfg.Enabled {
		return nil
	}

	if cfg.Topic == "" {
		return fmt.Errorf("APNs topic is required when APNs is enabled")
	}

	// Check if either token-based or certificate-based auth is configured
	hasTokenAuth := cfg.KeyID != "" && cfg.TeamID != "" && cfg.KeyFile != ""
	hasCertAuth := cfg.CertFile != ""

	if !hasTokenAuth && !hasCertAuth {
		return fmt.Errorf("either token-based (key_id, team_id, key_file) or certificate-based (cert_file) authentication is required")
	}

	// Validate files exist
	if hasTokenAuth {
		if _, err := os.Stat(cfg.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("APNs key file does not exist: %s", cfg.KeyFile)
		}
	}

	if hasCertAuth {
		if _, err := os.Stat(cfg.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("APNs certificate file does not exist: %s", cfg.CertFile)
		}
	}

	return nil
}