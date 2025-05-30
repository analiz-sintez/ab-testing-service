package models

import (
	"time"
)

type ConditionType string

const (
	ConditionTypeHeader    ConditionType = "header"
	ConditionTypeQuery     ConditionType = "query"
	ConditionTypeCookie    ConditionType = "cookie"
	ConditionTypeUserAgent ConditionType = "user_agent"
	ConditionTypeLanguage  ConditionType = "language"
	ConditionTypeExpr      ConditionType = "expr"
)

func (ct ConditionType) IsValid() bool {
	switch ct {
	case ConditionTypeHeader, ConditionTypeQuery, ConditionTypeCookie,
		ConditionTypeUserAgent, ConditionTypeLanguage, ConditionTypeExpr:
		return true
	}
	return false
}

// RouteCondition represents a condition for routing traffic
type RouteCondition struct {
	Type      ConditionType     `json:"type" db:"type"`           // Type of condition: "header", "query", "cookie", "user_agent", "language", "expr"
	ParamName string            `json:"param_name" db:"param"`    // Name of the parameter to check (for header, query, cookie) or expression for expr type
	Values    map[string]string `json:"values" db:"values"`       // List of values to match targets by id or expressions for expr type
	Default   string            `json:"default" db:"default"`     // Default target ID if no match is found
	Expr      string            `json:"expr,omitempty" db:"expr"` // Expression for expr type condition
}

type Target struct {
	ID       string  `json:"id" db:"id"`
	Name     string  `json:"name" db:"name"`
	URL      string  `json:"url" db:"url"`
	Weight   float64 `json:"weight" db:"weight"`
	IsActive bool    `json:"is_active" db:"is_active"`
	ProxyID  string  `json:"proxy_id" db:"proxy_id"`
}

type Visit struct {
	ID        string    `json:"id" db:"id"`
	ProxyID   string    `json:"proxy_id" db:"proxy_id"`
	TargetID  string    `json:"target_id" db:"target_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	RID       string    `json:"rid" db:"rid"`
	RRID      string    `json:"rrid" db:"rrid"`
	RUID      string    `json:"ruid" db:"ruid"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
