package handlers

import (
	"fmt"
	"mcp-google-search-server/internal/models"
	"mcp-google-search-server/internal/services"
	"strings"

	"github.com/sirupsen/logrus"
)

type SearchHandler struct {
	googleSearch *services.GoogleSearchService
}

func NewSearchHandler(googleSearch *services.GoogleSearchService) *SearchHandler {
	return &SearchHandler{
		googleSearch: googleSearch,
	}
}

func (h *SearchHandler) HandleCallTool(params models.CallToolParams) (models.CallToolResult, error) {
	switch params.Name {
	case "google_search":
		return h.handleGoogleSearch(params.Arguments)
	default:
		return models.CallToolResult{}, fmt.Errorf("unknown tool: %s", params.Name)
	}
}

func (h *SearchHandler) handleGoogleSearch(args map[string]interface{}) (models.CallToolResult, error) {
	query, ok := args["query"].(string)
	if !ok {
		return models.CallToolResult{}, fmt.Errorf("query parameter is required and must be a string")
	}

	if strings.TrimSpace(query) == "" {
		return models.CallToolResult{}, fmt.Errorf("query cannot be empty")
	}

	logrus.Debugf("Performing Google search for query: %s", query)

	searchResp, err := h.googleSearch.Search(query)
	if err != nil {
		logrus.Errorf("Google search failed: %v", err)
		return models.CallToolResult{}, fmt.Errorf("search failed: %w", err)
	}

	// Format results as text
	var resultText strings.Builder
	resultText.WriteString(fmt.Sprintf("Search results for: %s\n\n", query))

	if len(searchResp.Items) == 0 {
		resultText.WriteString("No results found.")
	} else {
		resultText.WriteString(fmt.Sprintf("Found %d results:\n\n", len(searchResp.Items)))
		for i, item := range searchResp.Items {
			resultText.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title))
			resultText.WriteString(fmt.Sprintf("   URL: %s\n", item.Link))
			resultText.WriteString(fmt.Sprintf("   %s\n\n", item.Snippet))
		}
	}

	return models.CallToolResult{
		Content: []models.ToolContent{
			{
				Type: "text",
				Text: resultText.String(),
			},
		},
	}, nil
}

func (h *SearchHandler) GetTools() []models.Tool {
	return []models.Tool{
		{
			Name:        "google_search",
			Description: "Search the web using Google Custom Search API",
			InputSchema: models.InputSchema{
				Type: "object",
				Properties: map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query",
					},
				},
				Required: []string{"query"},
			},
		},
	}
}
