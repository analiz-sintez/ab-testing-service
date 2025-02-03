// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package storage

import (
	"context"
)

type Querier interface {
	CreateProxy(ctx context.Context, arg *CreateProxyParams) error
	CreateProxyChange(ctx context.Context, arg *CreateProxyChangeParams) error
	CreateTarget(ctx context.Context, arg *CreateTargetParams) error
	CreateUser(ctx context.Context, arg *CreateUserParams) error
	CreateVisit(ctx context.Context, arg *CreateVisitParams) error
	DeleteTargetByProxyID(ctx context.Context, proxyID string) error
	GetAllTags(ctx context.Context) ([]string, error)
	GetProxies(ctx context.Context) ([]*GetProxiesRow, error)
	GetProxiesByTags(ctx context.Context, tags []string) ([]*GetProxiesByTagsRow, error)
	GetProxy(ctx context.Context, id string) (*GetProxyRow, error)
	GetProxyChangesByProxyID(ctx context.Context, arg *GetProxyChangesByProxyIDParams) ([]*ProxyChange, error)
	GetProxyTags(ctx context.Context, id string) ([]string, error)
	GetStats(ctx context.Context, arg *GetStatsParams) (*GetStatsRow, error)
	GetTargetStats(ctx context.Context, arg *GetTargetStatsParams) ([]*GetTargetStatsRow, error)
	GetTargetsByProxyID(ctx context.Context, proxyID string) ([]*GetTargetsByProxyIDRow, error)
	GetUniqueUsersCount(ctx context.Context, arg *GetUniqueUsersCountParams) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateProxyCondition(ctx context.Context, arg *UpdateProxyConditionParams) error
	UpdateProxyTags(ctx context.Context, arg *UpdateProxyTagsParams) error
	UserExists(ctx context.Context, email string) (bool, error)
}

var _ Querier = (*Queries)(nil)
