package service

import (
	"gorm.io/gorm"
	"katydid-mp-user/utils"
)

// TODO:GG behaviour , 过滤器，依赖注入

type Ctx struct {
	ActorId    uint64      // 操作者ID
	ActorType  uint8       // 操作者类型
	Permission []string    // 权限列表 TODO:GG 权限Casbin?
	Extra      utils.KSMap // 扩展信息
	Tx         *gorm.DB    // 事务对象
}

func NewCtx(
	actorId uint64, actorType uint8,
	extra utils.KSMap,
) *Ctx {
	if extra == nil {
		extra = make(map[string]any)
	}
	return &Ctx{
		ActorId:    actorId,
		ActorType:  actorType,
		Permission: nil,
		Extra:      extra,
		Tx:         nil,
	}
}

type IService[T any] interface {
	WithCtx(ctx *Ctx) IService[T]
	Ctx() *Ctx
	WithTx(tx *gorm.DB) IService[T]
	//Behaviour() IService
	//Add()
	//Del()
	//Upd()
	//Get()
}

type Base struct {
	ctx *Ctx
}

func NewBase(ctx *Ctx) *Base {
	return &Base{ctx: ctx}
}

func (s *Base) Ctx() *Ctx {
	return s.ctx
}

func (s *Base) WithCtx(ctx *Ctx) IService[any] {
	s.ctx = ctx
	return s
}

func (s *Base) WithTx(tx *gorm.DB) IService[any] {
	s.ctx.Tx = tx
	return s
}
