package proxy

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"

	"github.com/ab-testing-service/internal/models"
)

func (p *Proxy) selectTarget(r *http.Request) (*Target, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// First, try to get target from cookie
	if target := p.getTargetFromCookie(r); target != nil {
		return target, nil
	}

	// Then, check routing conditions if present
	if p.Config.Condition != nil {
		if target := p.getTargetByCondition(r); target != nil {
			return target, nil
		}
		return nil, fmt.Errorf("no matching target found")
	}

	// Fall back to weighted random selection if no condition matches or no condition is set
	var totalWeight float64
	var activeTargets []Target

	for _, target := range p.Targets {
		if target.IsActive {
			totalWeight += target.Weight
			activeTargets = append(activeTargets, target)
		}
	}

	if len(activeTargets) == 0 {
		return nil, fmt.Errorf("no active targets available")
	}

	// Choose a random target based on weights
	rnd := rand.Float64() * totalWeight
	var cumulativeWeight float64

	for _, target := range activeTargets {
		cumulativeWeight += target.Weight
		if rnd <= cumulativeWeight {
			return &target, nil
		}
	}

	// Fallback to the first active target if something goes wrong with the random selection
	return &activeTargets[0], nil
}

func (p *Proxy) getTargetByCondition(r *http.Request) *Target {
	var value string

	if p.Config.Condition == nil {
		log.Printf("No condition defined for proxy %s, using default target selection", p.ID)
		// If no condition is defined, return nil to fall back to weighted random selection
		return nil
	}

	if p.Config.Condition.Type == models.ConditionTypeExpr {
		// Use expr-lang/expr to evaluate the expression
		log.Printf("Using expression-based condition for proxy %s", p.ID)
		return p.getTargetByExpr(r)
	}

	switch p.Config.Condition.Type {
	case models.ConditionTypeHeader:
		value = r.Header.Get(p.Config.Condition.ParamName)
		log.Printf("Header condition: %s=%s for proxy %s", p.Config.Condition.ParamName, value, p.ID)
	case models.ConditionTypeQuery:
		value = r.URL.Query().Get(p.Config.Condition.ParamName)
		log.Printf("Query condition: %s=%s for proxy %s", p.Config.Condition.ParamName, value, p.ID)
	case models.ConditionTypeCookie:
		cookie, err := r.Cookie(p.Config.Condition.ParamName)
		if err == nil {
			value = cookie.Value
			log.Printf("Cookie condition: %s=%s for proxy %s", p.Config.Condition.ParamName, value, p.ID)
		} else {
			log.Printf("Cookie %s not found for proxy %s", p.Config.Condition.ParamName, p.ID)
		}
	case models.ConditionTypeUserAgent:
		ua := r.Header.Get("User-Agent")
		switch p.Config.Condition.ParamName {
		case "platform":
			value = detectPlatform(ua)
			log.Printf("User-Agent platform condition: %s for proxy %s", value, p.ID)
		case "browser":
			value = detectBrowser(ua)
			log.Printf("User-Agent browser condition: %s for proxy %s", value, p.ID)
		default:
			log.Printf("Unknown User-Agent parameter: %s for proxy %s", p.Config.Condition.ParamName, p.ID)
		}
	case models.ConditionTypeLanguage:
		value = parseAcceptLanguage(r.Header.Get("Accept-Language"))
		log.Printf("Language condition: %s for proxy %s", value, p.ID)
	default:
		log.Printf("Unknown condition type: %s for proxy %s, using default target", p.Config.Condition.Type, p.ID)
		return p.getTargetById(p.Config.Condition.Default)
	}

	// Check if the value matches any of the specified values
	if targetID, ok := p.Config.Condition.Values[value]; ok {
		log.Printf("Condition value %s matched target %s for proxy %s", value, targetID, p.ID)
		if target := p.getTargetById(targetID); target != nil {
			return target
		}
		log.Printf("Target %s not found for proxy %s", targetID, p.ID)
	} else {
		log.Printf("No match for condition value %s in proxy %s, using default target", value, p.ID)
	}

	// Fall back to default target
	defaultTarget := p.getTargetById(p.Config.Condition.Default)
	if defaultTarget == nil {
		log.Printf("Default target %s not found for proxy %s", p.Config.Condition.Default, p.ID)
	}
	return defaultTarget
}

func (p *Proxy) getTargetById(id string) *Target {
	for _, target := range p.Targets {
		if target.ID == id && target.IsActive {
			return &target
		}
	}
	return nil
}

// detectPlatform detects the platform (mobile/desktop) from User-Agent
func detectPlatform(ua string) string {
	ua = strings.ToLower(ua)
	mobileKeywords := []string{
		"mobile", "android", "iphone", "ipad", "ipod",
		"windows phone", "blackberry", "opera mini",
	}

	for _, keyword := range mobileKeywords {
		if strings.Contains(ua, keyword) {
			return "mobile"
		}
	}
	return "desktop"
}

// detectBrowser detects the browser from User-Agent
func detectBrowser(ua string) string {
	ua = strings.ToLower(ua)
	browsers := map[string]string{
		"firefox":   "firefox",
		"chrome":    "chrome",
		"safari":    "safari",
		"edge":      "edge",
		"opera":     "opera",
		"msie":      "ie",
		"trident/7": "ie",
	}

	for keyword, browser := range browsers {
		if strings.Contains(ua, keyword) {
			return browser
		}
	}
	return "other"
}

// parseAcceptLanguage gets the preferred language from Accept-Language header
func parseAcceptLanguage(acceptLang string) string {
	if acceptLang == "" {
		return ""
	}

	// Split the Accept-Language header into parts
	parts := strings.Split(acceptLang, ",")
	if len(parts) == 0 {
		return ""
	}

	// Get the first (most preferred) language
	lang := strings.Split(parts[0], ";")[0]
	return strings.ToLower(strings.TrimSpace(lang))
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check for X-Real-IP header
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
