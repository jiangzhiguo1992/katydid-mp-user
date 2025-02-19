package model

type (
	Base struct {
		//gorm.Model
		//IDBModel

		Id       uint64 `json:"id" gorm:"primarykey"`
		CreateAt int64  `json:"createAt" gorm:"autoCreateTime:milli"`
		UpdateAt int64  `json:"updateAt" gorm:"autoUpdateTime:milli"`

		// TODO:GG 所有的查询都带上index `gorm:"index"`
		DeleteBy int64  `json:"deleteBy"`
		DeleteAt *int64 `json:"deleteAt"`

		Extra map[string]interface{} `json:"extra" gorm:"serializer:json"` // 额外信息
	}

	//IDBModel interface {
	//	CheckFields() []*tools.CodeError
	//}
)

func NewBaseEmpty() *Base {
	return &Base{
		//Id: nil, // auto
		//CreateAt: time.Now().UnixMilli(), // auto
		//UpdateAt: time.Now().UnixMilli(), // auto
		DeleteBy: 0,
		DeleteAt: nil,
		Extra:    map[string]interface{}{},
	}
}

const (
	deleteByUserOffset  = 10000  // 用户删除偏移量
	deleteByUserSelf    = 1      // 用户自己删除
	deleteByAdminSys    = -1     // 系统管理员删除
	deleteByAdminOffset = -10000 // 系统管理员删除偏移量
)

func (b *Base) GetDelByUserSelf() int64 {
	return deleteByUserSelf
}

func (b *Base) GetDelByAdminSys() int64 {
	return deleteByAdminSys
}

func (b *Base) GetDelBy(id uint64) int64 {
	by := int64(id)
	if b.IsDelByUser() {
		by = by + deleteByUserOffset
	} else if b.IsDelByAdmin() {
		by = -by + deleteByAdminOffset
	}
	return by
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
