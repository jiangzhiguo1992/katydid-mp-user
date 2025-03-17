package model

import (
	"katydid-mp-user/pkg/data"
	"katydid-mp-user/pkg/valid"
	"time"
)

type (
	// Base 基础结构体
	Base struct {
		//gorm.Model
		ID uint64 `json:"id" gorm:"primarykey"` // 主键 TODO:GG 雪花限时太多了，应该只考虑分布式

		Status   Status `json:"status" gorm:"default:0"`              // 状态
		CreateAt int64  `json:"createAt" gorm:"autoCreateTime:milli"` // 创建时间
		UpdateAt int64  `json:"updateAt" gorm:"autoUpdateTime:milli"` // 更新时间
		DeleteAt *int64 `json:"deleteAt"`                             // 删除时间 // TODO:GG 所有的查询都带上index `gorm:"index"`
		DeleteBy int64  `json:"deleteBy"`                             // 删除人

		// id
		// index
		// required
		// stats

		Extra data.KSMap `json:"extra" gorm:"serializer:json"` // 额外信息 (!索引/!必需)
	}

	// Status 状态 (组合体可自定义)
	Status int
)

func NewBaseEmpty() *Base {
	return &Base{
		//ID: nil, // auto
		Status: StatusInit, // auto
		//CreateAt: time.Now().UnixMilli(), // auto
		//UpdateAt: time.Now().UnixMilli(), // auto
		DeleteBy: 0,
		DeleteAt: nil,
		Extra:    make(data.KSMap),
	}
}

func NewBase(extra data.KSMap) *Base {
	return &Base{
		//ID: nil, // auto
		Status: StatusInit, // auto
		//CreateAt: time.Now().UnixMilli(), // auto
		//UpdateAt: time.Now().UnixMilli(), // auto
		DeleteBy: 0,
		DeleteAt: nil,
		Extra:    extra,
	}
}

func (b *Base) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll:        valid.FieldValidRule{},
		valid.SceneBind:       valid.FieldValidRule{},
		valid.SceneSave:       valid.FieldValidRule{},
		valid.SceneInsert:     valid.FieldValidRule{},
		valid.SceneUpdate:     valid.FieldValidRule{},
		valid.SceneQuery:      valid.FieldValidRule{},
		valid.SceneReturn:     valid.FieldValidRule{},
		valid.SceneCustom + 1: valid.FieldValidRule{},
	}
}

func (b *Base) ValidExtraRules() (data.KSMap, valid.ExtraValidRules) {
	return b.Extra, valid.ExtraValidRules{
		valid.SceneAll: map[valid.Tag]valid.ExtraValidRuleInfo{
			// 管理员备注 (0-10000)
			extraKeyAdminNote: {
				Field: extraKeyAdminNote,
				ValidFn: func(value any) bool {
					val, ok := value.(string)
					if !ok {
						return false
					}
					return len(val) <= 10_000
				},
			},
		},
	}
}

func (b *Base) ValidStructRules(scene valid.Scene, fn valid.FuncReportError) {
	switch scene {
	case valid.SceneQuery:
		if b.CreateAt < b.UpdateAt {
			fn(b.CreateAt, "CreateAt", valid.TagFormat, "")
		}
		if (b.DeleteAt == nil) && (b.DeleteBy != 0) {
			fn(b.DeleteAt, "DeleteAt", valid.TagFormat, "")
		} else if (b.DeleteAt != nil) && (b.DeleteBy == 0) {
			fn(b.DeleteBy, "DeleteBy", valid.TagFormat, "")
		}
	}
}

func (b *Base) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagFormat: {
					"CreateAt": {"format_create_at_err", false, nil},
					"DeleteAt": {"format_delete_at_err", false, nil},
					"DeleteBy": {"format_delete_by_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				extraKeyAdminNote: {"format_admin_note_err", false, nil},
			},
		},
	}
}

const (
	StatusBlack Status = -1 // 黑名单
	StatusInit  Status = 0  // 初始
	StatusWhite Status = 1  // 白名单
)

const (
	deleteByUserOffset  = 10000  // 用户删除偏移量
	deleteByUserSelf    = 1      // 用户自己删除
	deleteByAdminSys    = -1     // 系统管理员删除
	deleteByAdminOffset = -10000 // 系统管理员删除偏移量
)

func (b *Base) IsBlack() bool {
	return b.Status <= StatusBlack
}

func (b *Base) IsWhite() bool {
	return b.Status >= StatusWhite
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

const (
	extraKeyAdminNote = "adminNote" // 管理员备注
)

func (b *Base) GetAdminNote() (string, bool) {
	return b.Extra.GetString(extraKeyAdminNote)
}

func (b *Base) SetAdminNote(adminNote *string) {
	b.Extra.SetString(extraKeyAdminNote, adminNote)
}
