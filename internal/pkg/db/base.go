package db

import "gorm.io/gorm"

type (
	// Base 基础数据库
	Base struct {
		W  *gorm.DB // 写数据库
		R  *gorm.DB // 读数据库
		Tx *gorm.Tx // 事务数据库
	}
)

func NewBase() *Base {
	return &Base{
		W:  nil,
		R:  nil,
		Tx: nil,
	}
}
