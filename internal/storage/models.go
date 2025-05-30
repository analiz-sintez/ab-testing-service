// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package storage

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type ProxyChange struct {
	ID            string
	ProxyID       string
	ChangeType    string
	PreviousState []byte
	NewState      []byte
	CreatedAt     pgtype.Timestamptz
	CreatedBy     *string
}

type ProxyListenUrl struct {
	ID        string
	ProxyID   string
	ListenUrl string
	PathKey   *string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
}
