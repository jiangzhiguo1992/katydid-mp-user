package model

import (
	"errors"
	"katydid-mp-user/internal/pkg/text"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/utils"
)

const (
	// 删除相关常量
	deleteByUserOffset  = 10000  // 用户删除偏移量
	deleteByUserSelf    = 1      // 用户自己删除
	deleteByAdminSys    = -1     // 系统管理员删除
	deleteByAdminOffset = -10000 // 系统管理员删除偏移量

	// Extra 相关常量
	extraKeyAdminNote = "adminNote" // 管理员备注
	extraItemMaxLen   = 10000
)

type (
	IValid interface {
		Valid() *err.CodeErrs
	}

	Base struct {
		//gorm.Model

		Id       uint64 `json:"id" gorm:"primarykey"`
		CreateAt int64  `json:"createAt" gorm:"autoCreateTime:milli"`
		UpdateAt int64  `json:"updateAt" gorm:"autoUpdateTime:milli"`
		DeleteAt *int64 `json:"deleteAt"` // TODO:GG 所有的查询都带上index `gorm:"index"`
		DeleteBy int64  `json:"deleteBy"`

		Extra utils.KSMap `json:"extra" gorm:"serializer:json"` // 额外信息

		ExtraKeys func() []string `json:"-" gorm:"-:all"`
	}
)

func NewBaseEmpty() *Base {
	return &Base{
		//Id: nil, // auto
		//CreateAt: time.Now().UnixMilli(), // auto
		//UpdateAt: time.Now().UnixMilli(), // auto
		DeleteBy: 0,
		DeleteAt: nil,
		Extra:    map[string]any{},
	}
}

func (b *Base) GetDelByUserSelf() int64 {
	return deleteByUserSelf
}

func (b *Base) GetDelByAdminSys() int64 {
	return deleteByAdminSys
}

func (b *Base) GetDelBy(id uint64) int64 {
	switch {
	case b.IsDelByUser():
		return int64(id) + deleteByUserOffset
	case b.IsDelByAdmin():
		return -int64(id) + deleteByAdminOffset
	default:
		return int64(id)
	}
}

func (b *Base) IsDelByUser() bool {
	return b.DeleteBy >= deleteByUserSelf
}

func (b *Base) IsDelByAdmin() bool {
	return b.DeleteBy <= deleteByAdminSys
}

func (b *Base) IsDelByUserSelf() bool {
	return b.DeleteBy == b.GetDelByUserSelf()
}

func (b *Base) IsDelByAdminSys() bool {
	return b.DeleteBy == b.GetDelByAdminSys()
}

func (b *Base) Valid() *err.CodeErrs {
	var errs = new(err.CodeErrs)
	// extra
	if v, ok := b.Extra[extraKeyAdminNote]; ok {
		if len(v.(string)) > extraItemMaxLen {
			errs = errs.WrapErrs(errors.New(text.MsgIdDBFieldLarge))
		}
	}
	return errs
}
