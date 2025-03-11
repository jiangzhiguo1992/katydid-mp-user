package db

import "gorm.io/gorm"

type (
	IDB interface {
		WithRead(db *gorm.DB) IDB
		WithWrite(db *gorm.DB) IDB
		WithTx(tx *gorm.Tx) IDB
		//WithPageOffset(offset int) IDB
		//WithPageLimit(limit int) IDB
	}

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

func (b *Base) WithRead(db *gorm.DB) IDB {
	b.R = db
	return b
}

func (b *Base) WithWrite(db *gorm.DB) IDB {
	b.W = db
	return b
}

func (b *Base) WithTx(tx *gorm.Tx) IDB {
	b.Tx = tx
	return b
}
