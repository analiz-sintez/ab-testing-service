package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	redirectInfo, err := p.getOrCreateRedirectInfo(r)
	if err != nil {
		http.Error(w, "Failed to process redirect info", http.StatusInternalServerError)
		p.stats.IncrementErrors(p.ID, "")
		return
	}

	target, err := p.selectTarget(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to select target: %s", err), http.StatusInternalServerError)
		p.stats.IncrementErrors(p.ID, redirectInfo.RUID)
		return
	}
	log.Printf("Selected target: %s", target.URL)

	// Check if this is a redirect request from another proxy
	if r.Header.Get("X-Internal-Redirect") == "true" {
		// todo prevent redirect loops
		return
	}

	p.setCookies(w, redirectInfo)

	// Get user identifier (prefer X-User-ID header, fallback to IP)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = redirectInfo.RUID
	}

	// Track request with user ID
	p.stats.IncrementRequestsWithUser(target.ID, userID)

	defer func() {
		duration := time.Since(start).Seconds()
		p.metrics.LatencyHistogram.WithLabelValues(target.URL).Observe(duration)
		p.metrics.RequestsTotal.WithLabelValues(target.URL).Inc()
	}()

	// Check if the target URL has a different host
	//targetURL := p.appendRedirectParams(target.URL, redirectInfo)
	parsedTarget, err := url.Parse(target.URL)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	if parsedTarget.Scheme == "" {
		parsedTarget.Scheme = "https"
	}

	// If target host is different from current host, do external redirect
	if parsedTarget.Host != r.Host {
		log.Printf("Redirecting to %s", parsedTarget.String())
		http.Redirect(w, r, parsedTarget.String(), http.StatusMovedPermanently)
		return
	}

	// For same host, do internal redirect
	r.URL.Path = parsedTarget.Path
	r.URL.RawQuery = parsedTarget.RawQuery
	r.Header.Set("X-Internal-Redirect", "true")
	http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
}
