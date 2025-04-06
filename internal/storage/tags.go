package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ab-testing-service/internal/models"
)

func (s *Storage) UpdateProxyTags(ctx context.Context, proxyID string, tags []string) error {
	err := s.q.UpdateProxyTags(ctx, &UpdateProxyTagsParams{
		ID:   proxyID,
		Tags: tags,
	})
	if err != nil {
		return fmt.Errorf("failed to update proxy tags: %w", err)
	}

	return nil
}

func (s *Storage) GetAllTags(ctx context.Context) ([]string, error) {
	tags, err := s.q.GetAllTags(ctx)

	return tags, err
}

func (s *Storage) GetTags(ctx context.Context, proxyID string) ([]string, error) {
	tags, err := s.q.GetProxyTags(ctx, proxyID)

	return tags, err
}

func (s *Storage) GetProxiesByTags(ctx context.Context, tags []string) ([]*models.Proxy, error) {
	rows, err := s.q.GetProxiesByTags(ctx, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to query proxies by tags: %w", err)
	}

	var proxies []*models.Proxy

	for _, item := range rows {
		var conditionJSON *models.RouteCondition

		if len(item.Condition) > 0 {
			if err := json.Unmarshal(item.Condition, &conditionJSON); err != nil {
				return nil, fmt.Errorf("failed to unmarshal condition: %w", err)
			}
		}

		// Create proxy with basic info
		proxy := &models.Proxy{
			ID:               item.ID,
			Mode:             models.ProxyMode(item.Mode),
			Condition:        conditionJSON,
			Tags:             item.Tags,
			SavingCookiesFlg: item.SavingCookiesFlg,
			CreatedAt:        item.CreatedAt.Time,
			UpdatedAt:        item.UpdatedAt.Time,
		}

		// Set name if provided
		if item.Name != nil {
			proxy.Name = *item.Name
		}

		// Fetch listen URLs for this proxy
		listenURLRows, err := s.q.GetProxyListenURLs(ctx, item.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get listen URLs for proxy %s: %w", item.ID, err)
		}

		// Add listen URLs to proxy
		proxy.ListenURLs = make([]models.ListenURL, len(listenURLRows))
		for i, urlRow := range listenURLRows {
			proxy.ListenURLs[i] = models.ListenURL{
				ID:        urlRow.ID,
				ProxyID:   urlRow.ProxyID,
				ListenURL: urlRow.ListenUrl,
				PathKey:   urlRow.PathKey,
			}
		}

		proxies = append(proxies, proxy)
	}

	return proxies, nil
}
