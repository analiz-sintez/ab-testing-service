package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ab-testing-service/internal/models"
)

func (s *Storage) CreateProxy(ctx context.Context, proxy *models.Proxy) error {
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) (err error) {
		repo := New(tx)
		// Generate a proxy ID if not provided
		if proxy.ID == "" {
			proxy.ID = uuid.New().String()
		}

		// Marshal condition if present
		var conditionJSON []byte = nil
		if proxy.Condition != nil {
			bytes, err := json.Marshal(proxy.Condition)
			if err != nil {
				return fmt.Errorf("failed to marshal condition: %w", err)
			}
			conditionJSON = bytes
		}

		// Create proxy record
		now := time.Now()
		err = repo.CreateProxy(ctx, &CreateProxyParams{
			ID:                 proxy.ID,
			Name:               &proxy.Name,
			Mode:               string(proxy.Mode),
			Condition:          conditionJSON,
			Tags:               proxy.Tags,
			SavingCookiesFlg:   proxy.SavingCookiesFlg,
			QueryForwardingFlg: proxy.QueryForwardingFlg,
			CreatedAt:          pgtype.Timestamptz{Time: now},
			UpdatedAt:          pgtype.Timestamptz{Time: now},
		})

		if err != nil {
			return fmt.Errorf(
				"failed to insert proxy: %w, condition: %s, condition type: %T, condition value: %+v",
				err, conditionJSON, proxy.Condition, proxy.Condition,
			)
		}

		// Insert listen URLs
		for i := range proxy.ListenURLs {
			listenURL := &proxy.ListenURLs[i]

			// Generate ID for listen URL if not provided
			if listenURL.ID == "" {
				listenURL.ID = uuid.New().String()
			}

			// Ensure proxy ID is set
			listenURL.ProxyID = proxy.ID

			err = repo.CreateProxyListenURL(ctx, &CreateProxyListenURLParams{
				ID:        listenURL.ID,
				ProxyID:   listenURL.ProxyID,
				ListenUrl: listenURL.ListenURL,
				PathKey:   listenURL.PathKey,
				CreatedAt: pgtype.Timestamptz{Time: now},
				UpdatedAt: pgtype.Timestamptz{Time: now},
			})
			if err != nil {
				return fmt.Errorf("failed to insert listen URL: %w", err)
			}
		}

		// Insert targets
		for i := range proxy.Targets {
			target := &proxy.Targets[i]

			// Ensure proxy ID is set
			target.ProxyID = proxy.ID

			// Generate ID for target if not provided
			if target.ID == "" {
				target.ID = uuid.New().String()
			}

			err = repo.CreateTarget(ctx, &CreateTargetParams{
				ID:       target.ID,
				ProxyID:  target.ProxyID,
				Url:      target.URL,
				Weight:   target.Weight,
				IsActive: target.IsActive,
			})
			if err != nil {
				return fmt.Errorf("failed to insert target: %w", err)
			}
		}

		return nil
	})

	return err
}
