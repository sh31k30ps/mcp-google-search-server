package services

import (
	"context"
	"fmt"
	"mcp-google-search-server/internal/models"
	"strconv"

	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

type GoogleSearchService struct {
	service    *customsearch.Service
	searchID   string
	maxResults int
}

func NewGoogleSearchService(apiKey, searchID string, maxResults int) (*GoogleSearchService, error) {
	service, err := customsearch.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Search service: %w", err)
	}

	return &GoogleSearchService{
		service:    service,
		searchID:   searchID,
		maxResults: maxResults,
	}, nil
}

func (g *GoogleSearchService) Search(query string) (*models.SearchResponse, error) {
	call := g.service.Cse.List().
		Q(query).
		Cx(g.searchID).
		Num(int64(g.maxResults))

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var results []models.SearchResult
	for _, item := range resp.Items {
		results = append(results, models.SearchResult{
			Title:       item.Title,
			Link:        item.Link,
			Snippet:     item.Snippet,
			DisplayLink: item.DisplayLink,
		})
	}

	total := 0
	if resp.SearchInformation != nil {
		if total, err = strconv.Atoi(resp.SearchInformation.TotalResults); err != nil {
			return nil, err
		}
	}

	return &models.SearchResponse{
		Items: results,
		Query: query,
		Total: total,
	}, nil
}
