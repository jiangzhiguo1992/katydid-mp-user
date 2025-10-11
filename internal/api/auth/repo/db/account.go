package db

import (
	"katydid-mp-user/internal/api/auth/repo/storage"
)

// Account is a type alias for storage.Account
type Account = storage.Account

// NewAccount creates a new Account instance
func NewAccount() *Account {
	return storage.NewAccount()
}
