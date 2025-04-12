package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ab-testing-service/internal/models"
	"github.com/ab-testing-service/internal/proxy"
)

type Storage struct {
	q     Querier
	db    *pgxpool.Pool
	Redis *redis.Client
}

const (
	proxyTTL  = 1 * time.Hour
	targetTTL = 1 * time.Hour
)

func NewStorage(conn *pgxpool.Pool, redis *redis.Client) *Storage {
	return &Storage{
		q:     New(conn),
		db:    conn,
		Redis: redis,
	}
}

func (s *Storage) SaveVisit(ctx context.Context, visit *models.Visit) error {
	visit.ID = uuid.New().String()
	visit.CreatedAt = time.Now()

	_, err := s.db.Exec(ctx,
		`INSERT INTO visits (id, proxy_id, target_id, user_id, rid, rrid, ruid, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		visit.ID, visit.ProxyID, visit.TargetID, visit.UserID,
		visit.RID, visit.RRID, visit.RUID, visit.CreatedAt,
	)
	return err
}

func (s *Storage) GetProxies(ctx context.Context) ([]proxy.Config, error) {
	var proxies []proxy.Config
	rows, err := s.db.Query(ctx,
		`SELECT id, name, mode, condition, tags, saving_cookies_flg, query_forwarding_flg
		FROM proxies ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query proxies: %w", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate proxies: %w", err)
	}

	for rows.Next() {
		// failed to scan proxy: can't scan into dest[1]: cannot scan NULL into *string
		var p models.Proxy
		var conditionJSON []byte
		var name string
		if err := rows.Scan(&p.ID, &name, &p.Mode, &conditionJSON, &p.Tags, &p.SavingCookiesFlg, &p.QueryForwardingFlg); err != nil {
			return nil, fmt.Errorf("failed to scan proxy: %w", err)
		}
		if len(conditionJSON) > 0 {
			p.Condition = &models.RouteCondition{}
			if err := json.Unmarshal(conditionJSON, p.Condition); err != nil {
				return nil, fmt.Errorf("failed to unmarshal condition: %w", err)
			}
		}

		config := proxy.Config{
			ID:                   p.ID,
			Name:                 name,
			Mode:                 p.Mode,
			Tags:                 p.Tags,
			SavingCookiesFlg:     p.SavingCookiesFlg,
			QueryForwardingFlg:   p.QueryForwardingFlg,
			CookiesForwardingFlg: p.CookiesForwardingFlg,
		}

		// Fetch ListenURLs from proxy_listen_urls table
		listenURLs, err := s.q.GetProxyListenURLs(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get listen URLs for proxy %s: %w", p.ID, err)
		}

		// Map ListenURLs to the models.Proxy and proxy.Config structures
		for _, listenURL := range listenURLs {
			// Add to models.Proxy
			modelListenURL := models.ListenURL{
				ID:        listenURL.ID,
				ProxyID:   listenURL.ProxyID,
				ListenURL: listenURL.ListenUrl,
				PathKey:   listenURL.PathKey,
			}
			p.ListenURLs = append(p.ListenURLs, modelListenURL)

			// Add to proxy.Config
			configListenURL := proxy.ListenURL{
				ID:        listenURL.ID,
				ListenURL: listenURL.ListenUrl,
				PathKey:   listenURL.PathKey,
			}
			config.ListenURLs = append(config.ListenURLs, configListenURL)
		}

		condition, err := convertCondition(p.Condition)
		if err != nil {
			// Error handling
			log.Printf("Failed to convert condition for proxy %s: %v", p.ID, err)
			// Possible options:
			// 1. Skip this proxy
			//continue
			// 2. Return error
			//return nil, fmt.Errorf("failed to process proxy %s: %w", p.ID, err)
			// 3. Return nil and continue processing other proxies
			//config.Condition = nil
		}

		if condition != nil {
			config.Condition = condition
		}
		proxies = append(proxies, config)
	}
	return proxies, nil
}

// Безопасное приведение типов с обработкой ошибок
func convertCondition(rc *models.RouteCondition) (*proxy.Condition, error) {
	if rc == nil {
		return nil, nil // Если входной параметр nil, возвращаем nil без ошибки
	}

	// Проверяем поля на корректность
	if !rc.Type.IsValid() {
		return nil, fmt.Errorf("invalid condition type: %v", rc.Type)
	}

	// Создаем новый объект Condition
	condition := &proxy.Condition{
		Type:      rc.Type,
		ParamName: rc.ParamName,
		Values:    make(map[string]string),
		Default:   rc.Default,
		Expr:      rc.Expr,
	}

	// Копируем значения map, проверяя их валидность
	for k, v := range rc.Values {
		if k == "" {
			return nil, fmt.Errorf("empty key in Values map")
		}
		condition.Values[k] = v
	}

	return condition, nil
}
