package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

func (p *Proxy) appendRedirectParams(targetURL string, info *RedirectInfo) string {
	u, err := url.Parse(targetURL)
	if err != nil {
		return targetURL
	}

	// Get existing query parameters
	query := u.Query()

	// Add redirect info parameters
	query.Set("rid", info.RID)
	query.Set("rrid", info.RRID)
	query.Set("ruid", info.RUID)

	// Add all original query parameters if query forwarding is enabled
	if p.QueryForwardingFlg {
		for key, values := range info.Query {
			for _, value := range values {
				query.Add(key, value)
			}
		}
	}

	// Add cookies as query parameters if cookies forwarding is enabled
	if p.CookiesForwardingFlg {
		for _, cookie := range info.Cookies {
			query.Add(fmt.Sprintf("cookie_%s", cookie.Name), cookie.Value)
		}
	}

	// Set the updated query string
	u.RawQuery = query.Encode()

	return u.String()
}

func (p *Proxy) getOrCreateRedirectInfo(r *http.Request) (*RedirectInfo, error) {
	// Get or generate RUID from cookie
	ruidCookie, err := r.Cookie("ruid")
	var ruid string
	if errors.Is(err, http.ErrNoCookie) || ruidCookie == nil {
		ruid = uuid.New().String()
	} else {
		ruid = ruidCookie.Value
	}

	// Get RID from proxy ID
	rid := fmt.Sprintf("rid_%s", p.ID)

	// Generate new RRID for this request
	rrid := uuid.New().String()

	// Get original query parameters
	query := r.URL.Query()

	return &RedirectInfo{
		RID:   rid,
		RRID:  rrid,
		RUID:  ruid,
		Query: query,
	}, nil
}
