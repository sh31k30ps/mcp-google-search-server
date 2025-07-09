package mcp

import (
	"encoding/json"
	"fmt"
	"mcp-google-search-server/internal/config"
	"mcp-google-search-server/internal/handlers"
	"mcp-google-search-server/internal/models"
	"mcp-google-search-server/internal/services"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config        *config.Config
	searchHandler *handlers.SearchHandler
	upgrader      websocket.Upgrader
}

func NewServer(cfg *config.Config) *Server {
	googleSearch, err := services.NewGoogleSearchService(
		cfg.GoogleAPIKey,
		cfg.GoogleSearchID,
		cfg.MaxResults,
	)
	if err != nil {
		logrus.Fatalf("Failed to initialize Google Search service: %v", err)
	}

	return &Server{
		config:        cfg,
		searchHandler: handlers.NewSearchHandler(googleSearch),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"name":        "mcp-google-search-server",
		"version":     "1.0.0",
		"description": "MCP server for Google Search API",
		"status":      "running",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	logrus.Info("New WebSocket connection established")

	for {
		var msg models.MCPMessage
		if err := conn.ReadJSON(&msg); err != nil {
			logrus.Errorf("Failed to read WebSocket message: %v", err)
			break
		}

		response := s.handleMessage(msg)
		if err := conn.WriteJSON(response); err != nil {
			logrus.Errorf("Failed to write WebSocket response: %v", err)
			break
		}
	}
}

func (s *Server) handleMessage(msg models.MCPMessage) models.MCPMessage {
	logrus.Debugf("Handling message: %s", msg.Method)

	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "tools/list":
		return s.handleListTools(msg)
	case "tools/call":
		return s.handleCallTool(msg)
	default:
		return models.MCPMessage{
			ID:      msg.ID,
			JSONRPC: "2.0",
			Error: &models.MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", msg.Method),
			},
		}
	}
}

func (s *Server) handleInitialize(msg models.MCPMessage) models.MCPMessage {
	result := models.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: models.ServerCapabilities{
			Tools: map[string]any{},
		},
		ServerInfo: models.ServerInfo{
			Name:    "mcp-google-search-server",
			Version: "1.0.0",
		},
	}

	resultData, _ := json.Marshal(result)
	return models.MCPMessage{
		ID:      msg.ID,
		JSONRPC: "2.0",
		Result:  resultData,
	}
}

func (s *Server) handleListTools(msg models.MCPMessage) models.MCPMessage {
	result := models.ListToolsResult{
		Tools: s.searchHandler.GetTools(),
	}

	resultData, _ := json.Marshal(result)
	return models.MCPMessage{
		ID:      msg.ID,
		JSONRPC: "2.0",
		Result:  resultData,
	}
}

func (s *Server) handleCallTool(msg models.MCPMessage) models.MCPMessage {
	var params models.CallToolParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return models.MCPMessage{
			ID:      msg.ID,
			JSONRPC: "2.0",
			Error: &models.MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	result, err := s.searchHandler.HandleCallTool(params)
	if err != nil {
		return models.MCPMessage{
			ID:      msg.ID,
			JSONRPC: "2.0",
			Error: &models.MCPError{
				Code:    -32603,
				Message: err.Error(),
			},
		}
	}

	resultData, _ := json.Marshal(result)
	return models.MCPMessage{
		ID:      msg.ID,
		JSONRPC: "2.0",
		Result:  resultData,
	}
}
