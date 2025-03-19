package storage

import (
	"context"
	"gorm.io/gorm"
	"katydid-mp-user/pkg/storage"
)

// 表名常量定义
const (
	TableGroupAuth       TableName = "auths"
	TableAuthVerify                = TableGroupAuth + ".verify"
	TableAuthAccount               = TableGroupAuth + ".account"
	TableAuthAuth                  = TableGroupAuth + ".auth"
	TableAuthAccountAuth           = TableGroupAuth + ".account_auth"
	TableAuthToken                 = TableGroupAuth + ".token"
	TableAuthAccess                = TableGroupAuth + ".access"

	TableGroupUser TableName = "users"

	TableGroupClient TableName = "clients"

	TableGroupRole TableName = "roles"
)

type (
	// TableName 表名类型
	TableName string

	// Base 提供基础数据库功能
	Base struct {
		ctx context.Context
	}
)

// 默认连接名称
const (
	DefaultPsqlName = "default_psql"
	DefaultMsqlName = "default_mysql"
)

// NewBase 创建一个新的Base实例
func NewBase(ctx context.Context) *Base {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Base{
		ctx: ctx,
	}
}

// WithContext 设置上下文
func (b *Base) WithContext(ctx context.Context) *Base {
	return &Base{
		ctx: ctx,
	}
}

// Context 获取上下文
func (b *Base) Context() context.Context {
	if b.ctx == nil {
		return context.Background()
	}
	return b.ctx
}

// Psql 获取PostgreSQL连接
func (b *Base) Psql() *gorm.DB {
	return b.GetDB(DefaultPsqlName).WithContext(b.Context())
}

// Msql 获取MySQL连接
func (b *Base) Msql() *gorm.DB {
	return b.GetDB(DefaultMsqlName).WithContext(b.Context())
}

// GetDB 通过名称获取数据库连接
func (b *Base) GetDB(name string) *gorm.DB {
	return storage.GetDB(name)
}
