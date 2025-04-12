package models

import (
	"time"
)

type ProxyMode string

const (
	ProxyModeRedirect ProxyMode = "redirect"
	ProxyModePath     ProxyMode = "path"
)

type ListenURL struct {
	ID        string    `json:"id" db:"id"`
	ProxyID   string    `json:"proxy_id" db:"proxy_id"`
	ListenURL string    `json:"listen_url" db:"listen_url"`
	PathKey   *string   `json:"path_key,omitempty" db:"path_key"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Proxy struct {
	ID                   string          `json:"id" db:"id"`
	Name                 string          `json:"name" db:"name"`
	Mode                 ProxyMode       `json:"mode" db:"mode"`
	ListenURLs           []ListenURL     `json:"listen_urls"`
	Targets              []Target        `json:"targets" db:"targets"`
	Condition            *RouteCondition `json:"condition,omitempty" db:"condition"`
	Tags                 []string        `json:"tags" db:"tags"`
	IsActive             bool            `json:"is_active" db:"is_active"`
	SavingCookiesFlg     bool            `json:"saving_cookies_flg" db:"saving_cookies_flg"`
	QueryForwardingFlg   bool            `json:"query_forwarding_flg" db:"query_forwarding_flg"`
	CookiesForwardingFlg bool            `json:"cookies_forwarding_flg" db:"cookies_forwarding_flg"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}
