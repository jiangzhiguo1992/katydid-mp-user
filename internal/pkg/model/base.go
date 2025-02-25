package model

import (
	"katydid-mp-user/pkg/valid"
	"katydid-mp-user/utils"
	"time"
)

const (
	// 删除相关常量
	deleteByUserOffset  = 10000  // 用户删除偏移量
	deleteByUserSelf    = 1      // 用户自己删除
	deleteByAdminSys    = -1     // 系统管理员删除
	deleteByAdminOffset = -10000 // 系统管理员删除偏移量

	// Extra 相关常量
	extraKeyAdminNote       = "adminNote" // 管理员备注
	extraValAdminNoteMaxLen = 10000       // 最大长度
)

type (
	Base struct {
		//gorm.Model
		Id       uint64 `json:"id" gorm:"primarykey"`                 // 主键
		CreateAt int64  `json:"createAt" gorm:"autoCreateTime:milli"` // 创建时间
		UpdateAt int64  `json:"updateAt" gorm:"autoUpdateTime:milli"` // 更新时间
		DeleteAt *int64 `json:"deleteAt"`                             // 删除时间 // TODO:GG 所有的查询都带上index `gorm:"index"`
		DeleteBy int64  `json:"deleteBy"`                             // 删除人

		Extra utils.KSMap `json:"extra" gorm:"serializer:json"` // 额外信息 (!索引/!必需)
	}
)

func NewBaseEmpty() *Base {
	return &Base{
		//Id: nil, // auto
		//CreateAt: time.Now().UnixMilli(), // auto
		//UpdateAt: time.Now().UnixMilli(), // auto
		//DeleteBy: 0,
		//DeleteAt: nil,
		Extra: map[string]any{},
	}
}

func NewBaseDefault() *Base {
	return &Base{
		//Id: nil, // auto
		//CreateAt: time.Now().UnixMilli(), // auto
		//UpdateAt: time.Now().UnixMilli(), // auto
		DeleteBy: 0,
		DeleteAt: nil,
		Extra:    map[string]any{},
	}
}

func (b *Base) ValidFieldRules() valid.FieldValidResult {
	return valid.FieldValidResult{}
}

func (b *Base) ValidExtraRules() (utils.KSMap, valid.ExtraValidResult) {
	rules := map[string]valid.ExtraValidationInfo{
		// 管理员备注 (0-10000)
		extraKeyAdminNote: {
			Required: false,
			Validate: func(value interface{}) bool {
				str, ok := value.(string)
				if !ok {
					return false
				}
				return len(str) <= extraValAdminNoteMaxLen
			},
		},
	}
	return b.Extra, valid.ExtraValidResult{
		valid.SceneAll: rules,
	}
}

func (b *Base) ValidStructRules(_ any, _ valid.FuncReportError) {
}

func (b *Base) ValidRuleLocalizes() valid.RulesValidLocalize {
	return valid.RulesValidLocalize{}
}

func (b *Base) CreatedTime() time.Time {
	return time.UnixMilli(b.CreateAt)
}

func (b *Base) UpdatedTime() time.Time {
	return time.UnixMilli(b.UpdateAt)
}

func (b *Base) DeletedTime() *time.Time {
	if b.DeleteAt == nil {
		return nil
	}
	t := time.UnixMilli(*b.DeleteAt)
	return &t
}

func (b *Base) IsDeleted() bool {
	return b.DeleteAt != nil
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
