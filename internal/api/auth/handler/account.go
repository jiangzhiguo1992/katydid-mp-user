package handler

import (
	"katydid-mp-user/internal/api/auth/model"
	"katydid-mp-user/internal/api/auth/repo/cache"
	"katydid-mp-user/internal/api/auth/repo/db"
	"katydid-mp-user/internal/api/auth/service"
	"katydid-mp-user/internal/pkg/handler"
	"katydid-mp-user/pkg/data"
	"katydid-mp-user/pkg/errs"
)

type Account struct {
	*handler.Base
	service *service.Account
}

func NewAccount(
	db *db.Account, cache *cache.Account,
	getMaxNumByOwner func(ownKind model.TokenOwn, ownID uint64) (int, *errs.CodeErrs),
	isOwnerAuthEnable func(ownKind model.TokenOwn, ownID uint64, kind model.AuthKind) (bool, *errs.CodeErrs),
) *Account {
	return &Account{
		Base: handler.NewBase(nil),
		service: service.NewAccount(
			db, cache, isOwnerAuthEnable, getMaxNumByOwner,
		),
	}
}

func (a *Account) Post() {
	bind := &struct {
		OwnType  int     `json:"ownType" form:"ownType" binding:"required"`
		OwnId    uint64  `json:"ownId" form:"ownId" binding:"required"`
		Nickname string  `json:"nickname" form:"nickname"`
		UserId   *uint64 `json:"userId" form:"userId"`

		AuthKind int        `json:"authKind" form:"authKind" binding:"required"`
		Extra    data.KSMap `json:"extra" form:"extra" binding:"required"`
	}{}
	err := a.RequestBind(bind, true)
	if err != nil {
		a.Response400("", err)
		return
	}
	ownKind := model.TokenOwn(rune(bind.OwnType))
	//authKind := model.TokenOwn(rune(param.AuthKind))

	// TODO:GG 先check Verify？

	// TODO:GG 再Add Auth?

	add, err := a.service.AddAccount(ownKind, bind.OwnId, bind.UserId, bind.Nickname, nil)
	if err != nil {
		a.Response400("添加account失败", err)
		return
	}

	a.Response200(add)
}

func (a *Account) Del() {
	//name := c.Param("id")
	//u, err := strconv.ParseUint(name, 10, 64)
	//if err != nil {
	//	c.FString(http.StatusBadRequest, "Invalid Id")
	//	return
	//}
	//drop := c.DefaultPostForm("drop", "0")
	//ok := false
	//var codeError *utils.CodeError
	//if drop == "1" {
	//	ok, codeError = service.DropClient(u)
	//} else {
	//	// TODO:GG by
	//	ok, codeError = service.DelClient(u, 0)
	//}
	//if codeError != nil {
	//	c.JSON(http.StatusNotFound, codeError)
	//}
	//c.JSON(http.StatusOK, ok)
}

func (a *Account) Put() {

	//// TODO:GG 验证之后再设置密码，防止被别人注册但是没验证就知道密码了
	//
	//id := c.Param("id")
	//action := c.Param("action")
	//
	//// TODO:GG 先获取account
	//account := model.NewAccountEmpty()
	//
	//if action == "/avatarUrl" {
	//	// TODO:GG 修改avatarUrl
	//	avatarUrl := c.DefaultQuery("avatarUrl", "")
	//	account.SetAvatarUrl(&avatarUrl)
	//	// 修改url
	//	//ok, codeError := service.UpdClientAvatarUrl(u, avatarUrl)
	//	log.Debug("UpdAccount", log.FAny("avatarUrl", account))
	//}
	//
	//c.String(http.StatusOK, "account:%s by %s", id, action)
}

func (a *Account) Get() {
	//id := c.Param("id")
	//
	////firstname := c.DefaultQuery("firstname", "Guest")
	////lastname := c.Query("lastname")
	//
	//c.String(http.StatusOK, "account:%s", id)

	//name := c.Param("id")
	//u, err := strconv.ParseUint(name, 10, 64)
	//if err != nil {
	//	c.FString(http.StatusBadRequest, "Invalid Id")
	//	return
	//}
	//client, codeError := service.GetClient(u)
	//if codeError != nil {
	//	c.JSON(http.StatusNotFound, client)
	//}
	//c.JSON(http.StatusOK, client)
}
