package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ab-testing-service/internal/models"
	"github.com/ab-testing-service/internal/proxy"
)

type RouteCondition struct {
	Type      string   `json:"type" db:"type"`        // Type of condition: "header", "query", "cookie", "user_agent", "language", "expr"
	ParamName string   `json:"param_name" db:"param"` // Name of the parameter to check (for header, query, cookie)
	Values    []string `json:"values" db:"values"`    // List of parameter values to match targets
	Default   string   `json:"default" db:"default"`  // Default target ID if no match is found
	Expr      string   `json:"expr,omitempty"`        // Expression for expr type condition
}

// ExpressionConditionExample shows how to use expression-based conditions
// For example, to route traffic based on multiple conditions:
//
//	{
//	  "type": "expr",
//	  "expr": "headers[\"user-agent\"] contains \"iPhone\" && query[\"version\"] == \"2\"",
//	  "default": "target-id-1"
//	}
//
// Or using multiple expressions in the Values map:
//
//	{
//	  "type": "expr",
//	  "values": [
//	    "headers[\"user-agent\"] contains \"iPhone\"",
//	    "query[\"version\"] == \"2\""
//	  ],
//	  "default": "target-id-3"
//	}
//
// Available variables in expressions:
// - headers: map of request headers
// - query: map of query parameters
// - cookies: map of cookies
// - method: request method (GET, POST, etc.)
// - path: request path
// - host: request host
type CreateProxyRequest struct {
	ListenURL     string             `json:"listen_url" binding:"required"`
	ListenURLs    []string           `json:"listen_urls,omitempty"`
	Mode          string             `json:"mode" binding:"required"`
	Tags          []string           `json:"tags"`
	Targets       []CreateTargetSpec `json:"targets"`
	Condition     *RouteCondition    `json:"condition,omitempty"`
	PathKeyLength int                `json:"path_key_length,omitempty"` // Length of random path key for path-based routing
}

type CreateTargetSpec struct {
	URL      string  `json:"url" binding:"required"`
	Weight   float64 `json:"weight" binding:"required,min=0,max=1"`
	IsActive bool    `json:"is_active"`
}

type UpdateProxyURLRequest struct {
	ListenURL  string   `json:"listen_url"`
	ListenURLs []string `json:"listen_urls,omitempty"`
	PathKey    *string  `json:"path_key,omitempty"`
}

func generateRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(chars[rand.Intn(len(chars))])
	}
	return sb.String()
}

func (s *Server) createProxy(c *gin.Context) {
	var req CreateProxyRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default mode if not specified
	if req.Mode == "" {
		req.Mode = string(models.ProxyModeRedirect)
	}

	// Validate mode
	if req.Mode != string(models.ProxyModeRedirect) &&
		req.Mode != string(models.ProxyModePath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy mode"})
		return
	}

	// Create proxy model
	p := &models.Proxy{
		Mode: models.ProxyMode(req.Mode),
		Tags: req.Tags,
	}

	// Handle listen URLs
	if len(req.ListenURLs) > 0 {
		// Use the provided listen URLs array
		for _, url := range req.ListenURLs {
			if req.Mode == string(models.ProxyModePath) {
				if req.PathKeyLength == 0 {
					req.PathKeyLength = 8 // Default length
				}

				key := generateRandomString(req.PathKeyLength)
				// Create a ListenURL with the path key
				listenURL := models.ListenURL{
					ListenURL: url,
					PathKey:   &key,
				}
				p.ListenURLs = append(p.ListenURLs, listenURL)
			} else {
				// Create a ListenURL without path key
				listenURL := models.ListenURL{
					ListenURL: url,
				}
				p.ListenURLs = append(p.ListenURLs, listenURL)
			}
		}
	} else {
		// Backward compatibility: use the single listen URL
		if req.Mode == string(models.ProxyModePath) {
			if req.PathKeyLength == 0 {
				req.PathKeyLength = 8 // Default length
			}

			key := generateRandomString(req.PathKeyLength)
			// Create a ListenURL with the path key
			listenURL := models.ListenURL{
				ListenURL: req.ListenURL,
				PathKey:   &key,
			}
			p.ListenURLs = append(p.ListenURLs, listenURL)
		} else {
			// Create a ListenURL without path key
			listenURL := models.ListenURL{
				ListenURL: req.ListenURL,
			}
			p.ListenURLs = append(p.ListenURLs, listenURL)
		}
	}

	// Convert targets
	if len(req.Targets) > 0 {
		p.Targets = make([]models.Target, len(req.Targets))
		for i, t := range req.Targets {
			p.Targets[i] = models.Target{
				ID:       uuid.New().String(),
				URL:      t.URL,
				Weight:   t.Weight,
				IsActive: t.IsActive,
			}
		}
	}

	// Convert condition
	if req.Condition != nil && req.Condition.Type != "" {
		conditionType := models.ConditionType(req.Condition.Type)

		if !conditionType.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid condition type"})
			return
		}

		// For expression type, validate that we have either an Expr field or Values map
		if conditionType == models.ConditionTypeExpr {
			if req.Condition.Expr == "" && len(req.Condition.Values) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "expr condition requires either Expr field or Values map with expressions"})
				return
			}
		} else {
			// For non-expression types, validate that we have ParamName and Values
			if req.Condition.ParamName == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "param_name is required for non-expr conditions"})
				return
			}
			if len(req.Condition.Values) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "values are required for non-expr conditions"})
				return
			}
		}

		conditionValues := make(map[string]string, len(req.Condition.Values))
		for i, v := range req.Condition.Values {
			conditionValues[p.Targets[i].ID] = v
		}

		p.Condition = &models.RouteCondition{
			Type:      conditionType,
			ParamName: req.Condition.ParamName,
			Values:    conditionValues,
			Default:   req.Condition.Default,
			Expr:      req.Condition.Expr,
		}
	}

	// Create proxy in storage -> postgres
	if err := s.storage.CreateProxy(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create proxy in storage",
			"details": err.Error(),
		})
		return
	}

	// Create proxy configuration for supervisor
	cfg := proxy.Config{
		ID:   p.ID,
		Mode: p.Mode,
	}

	// Add ListenURLs to the config
	configListenURLs := make([]proxy.ListenURL, len(p.ListenURLs))
	for i, url := range p.ListenURLs {
		configListenURLs[i] = proxy.ListenURL{
			ListenURL: url.ListenURL,
			PathKey:   url.PathKey,
		}
	}
	cfg.ListenURLs = configListenURLs

	// Convert targets to config format
	if len(p.Targets) > 0 {
		cfg.Targets = make([]proxy.Target, len(p.Targets))
		for i, t := range p.Targets {
			cfg.Targets[i] = proxy.Target{
				ID:       t.ID,
				URL:      t.URL,
				Weight:   t.Weight,
				IsActive: t.IsActive,
			}
		}
	}

	// Add condition if provided
	if p.Condition != nil {
		cfg.Condition = &proxy.Condition{
			Type:      p.Condition.Type,
			ParamName: p.Condition.ParamName,
			Values:    p.Condition.Values,
			Default:   p.Condition.Default,
			Expr:      p.Condition.Expr,
		}
	}

	// Create proxy in supervisor -> start proxy server
	if err := s.supervisor.CreateProxy(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create proxy in supervisor",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (s *Server) updateProxyURL(c *gin.Context) {
	proxyID := c.Param("id")
	var req UpdateProxyURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	var userID *string
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			userID = &u.ID
		}
	}

	// Check if we're dealing with multiple listen URLs
	if len(req.ListenURLs) > 0 {
		// For now, use the first listen URL as the primary one
		// This is a temporary solution until we update the storage layer
		// to handle multiple listen URLs
		primaryURL := req.ListenURLs[0]

		// Update URL in storage with user ID
		if err := s.storage.UpdateProxyURL(c.Request.Context(), proxyID, primaryURL, req.PathKey, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// TODO: Add support for multiple listen URLs in the storage layer
		// This would involve creating a new method in the storage layer
		// to handle multiple listen URLs
	} else {
		// Update URL in storage with user ID (backward compatibility)
		if err := s.storage.UpdateProxyURL(c.Request.Context(), proxyID, req.ListenURL, req.PathKey, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Update supervisor
	cfg, err := s.storage.GetProxyConfig(c.Request.Context(), proxyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get updated proxy config: %v", err)})
		return
	}

	if err := s.supervisor.UpdateProxy(c.Request.Context(), cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update proxy in supervisor: %v", err)})
		return
	}

	c.Status(http.StatusOK)
}
