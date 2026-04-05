package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/accnotify/server/apns"
	"github.com/accnotify/server/config"
	"github.com/accnotify/server/handler"
	"github.com/accnotify/server/storage"
)

func main() {
	// Load configuration
	cfg := config.LoadFromEnv()

	// Initialize storage
	store, err := storage.NewSQLiteStorage(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize WebSocket hub
	hub := handler.NewHub(store)
	go hub.Run()

	// Initialize APNs service if enabled
	var apnsService *apns.Service
	apnsConfig := &apns.Config{
		Enabled:     cfg.APNsEnabled,
		CertFile:    cfg.APNsCertFile,
		KeyFile:     cfg.APNsKeyFile,
		KeyID:       cfg.APNsKeyID,
		TeamID:      cfg.APNsTeamID,
		Topic:       cfg.APNsTopic,
		Development: cfg.APNsDevelopment,
	}

	if cfg.APNsEnabled {
		if err := apns.ValidateConfig(apnsConfig); err != nil {
			log.Printf("APNs configuration validation failed: %v", err)
			log.Println("APNs service will be disabled")
		} else {
			apnsService, err = apns.NewService(apnsConfig)
			if err != nil {
				log.Printf("Failed to initialize APNs service: %v", err)
				log.Println("APNs service will be disabled")
			} else {
				log.Printf("APNs service initialized successfully (development: %v)", cfg.APNsDevelopment)
				defer apnsService.Close()
			}
		}
	} else {
		log.Println("APNs service is disabled")
	}

	// Initialize handlers
	pushHandler := handler.NewPushHandler(store, hub, apnsService)
	wsHandler := handler.NewWSHandler(hub, store)
	webhookHandler := handler.NewWebhookHandler(store, hub, apnsService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes - Prioritize static routes to avoid 404s
	router.GET("/ping", pushHandler.HandleHealth)
	router.GET("/health", pushHandler.HandleHealth)
	router.GET("/", pushHandler.HandleHealth)
	
	router.POST("/register", pushHandler.HandleRegister)
	router.GET("/register", pushHandler.HandleRegister) // Support Bark GET register
	
	router.POST("/push/:device_key", pushHandler.HandlePush)
	router.GET("/push/:device_key/*params", handleSimplePushParams(pushHandler))
	
	router.GET("/ws", func(c *gin.Context) {
		wsHandler.HandleConnect(c.Writer, c.Request)
	})

	// Finally, handle the greedy root-level route with a check for reserved words
	router.GET("/:device_key/*params", func(c *gin.Context) {
		key := c.Param("device_key")
		// If key is a reserved word, pass to the next handler (which will be a 404 if no other route matches)
		reserved := map[string]bool{
			"ping": true, "health": true, "register": true, "ws": true, "push": true, "webhook": true,
		}
		if reserved[key] {
			c.Next()
			return
		}
		handleSimplePushParams(pushHandler)(c)
	})

	// Webhook routes
	webhookGroup := router.Group("/webhook/:device_key")
	{
		webhookGroup.POST("", webhookHandler.HandleGenericWebhook)
		webhookGroup.POST("/github", webhookHandler.HandleGitHubWebhook)
		webhookGroup.POST("/gitlab", webhookHandler.HandleGitLabWebhook)
		webhookGroup.POST("/docker", webhookHandler.HandleDockerHubWebhook)
		webhookGroup.POST("/gitea", webhookHandler.HandleGiteaWebhook)
	}

	// Create server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Accnotify server starting on %s", addr)
		var err error
		if cfg.EnableHTTPS {
			err = srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// handleSimplePushParams handles /push/:device_key/*params
// Supports: /push/key/body OR /push/key/title/body OR /push/key/title/subtitle/body
func handleSimplePushParams(h *handler.PushHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := c.Param("params")
		// Remove leading slash
		if len(params) > 0 && params[0] == '/' {
			params = params[1:]
		}
		
		// Split by /
		parts := strings.Split(params, "/")
		
		var title, subtitle, body string
		switch len(parts) {
		case 1:
			// Only body provided
			title = "Accnotify"
			body = parts[0]
		case 2:
			// Title and body provided
			title = parts[0]
			body = parts[1]
		case 3:
			// Title, subtitle, and body provided
			title = parts[0]
			subtitle = parts[1]
			body = parts[2]
		default:
			if len(parts) > 3 {
				title = parts[0]
				subtitle = parts[1]
				// Join the rest as body
				body = strings.Join(parts[2:], "/")
			} else {
				title = "Accnotify"
				body = ""
			}
		}
		
		if subtitle != "" {
			body = subtitle + "\n" + body
		}
		
		c.Params = append(c.Params, gin.Param{Key: "title", Value: title})
		c.Params = append(c.Params, gin.Param{Key: "body", Value: body})
		h.HandleSimplePush(c)
	}
}
