package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mcp-google-search-server/internal/config"
	"mcp-google-search-server/pkg/mcp"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logging
	setupLogging(cfg.LogLevel)

	// Create MCP server
	mcpServer := mcp.NewServer(cfg)

	// Setup routes
	router := mux.NewRouter()
	router.HandleFunc("/", mcpServer.HandleRoot).Methods("GET")
	router.HandleFunc("/ws", mcpServer.HandleWebSocket).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	// Start server
	logrus.Infof("Starting MCP Google Search Server on port %s", cfg.Port)

	go func() {
		if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
			logrus.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")
}

func setupLogging(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
