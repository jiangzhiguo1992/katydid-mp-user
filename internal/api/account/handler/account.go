package handler

import (
	"github.com/gin-gonic/gin"
	"katydid-mp-user/internal/api/account/model"
	"katydid-mp-user/pkg/log"
	"net/http"
)

// TODO:GG 认证，手机/邮箱/用户名+密码
// TODO:GG 注册，设置密码、二级密码?

//type AccountPost struct {
//	AuthKind int16      `json:"auth_kind" form:"auth_kind" binding:"required"`
//	Extra    utils.KSMap `json:"extra" form:"extra" binding:"required"`
//	//Password string     `json:"password" form:"password" binding:"required,password"`
//	//Captcha string `json:"verify_code" form:"verify_code" binding:"required"` // TODO:GG 需要client来定
//}

func PostAccount(c *gin.Context) {
	//type aaa struct {
	//	AuthKind int16 `json:"auth_kind" form:"auth_kind" binding:"required"`
	//}
	//b:= &aaa{}
	//body := &utils.KSMap{}
	//if err := c.Bind(&body); err != nil {
	//	// 处理其他绑定错误
	//	params.Response(
	//		c,
	//		http.StatusBadRequest,
	//		"参数绑定失败",
	//		nil,
	//	)
	//	return
	//}
	//
	//// auth-kind
	//kind, ok := body.GetUint16("kind")
	//if !ok {
	//	params.Response(
	//		c,
	//		http.StatusBadRequest,
	//		"kind参数缺失",
	//		nil,
	//	)
	//	return
	//}
	//
	//// auth
	//auth, _ := service.GetAuth(kind, *body)
	//if auth != nil {
	//	params.Response(
	//		c,
	//		http.StatusConflict,
	//		"auth已存在",
	//		nil,
	//	)
	//	return
	//}
	//
	//// TODO:GG 需要先插入user吗?
	//
	//account, errors := service.AddAccount(0)
	//if errors != nil {
	//	params.Response(
	//		c,
	//		http.StatusInternalServerError,
	//		"添加account失败",
	//		errors,
	//	)
	//	return
	//}
	//
	//// TODO:GG 需要先插入auth吗?
	//
	//params.Response(
	//	c,
	//	http.StatusOK,
	//	"添加account成功",
	//	account,
	//)
}

type Response struct {
	gin.H
	//Status int         `json:"status"`
	//Msg    string      `json:"msg"`
	//Result any `json:"result"`
}

func DelAccount(c *gin.Context) {
	//name := c.Param("id")
	//u, err := strconv.ParseUint(name, 10, 64)
	//if err != nil {
	//	c.FString(http.StatusBadRequest, "Invalid ID")
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

func PutAccount(c *gin.Context) {

	// TODO:GG 验证之后再设置密码，防止被别人注册但是没验证就知道密码了

	id := c.Param("id")
	action := c.Param("action")

	// TODO:GG 先获取account
	account := model.NewAccountEmpty()

	if action == "/avatarUrl" {
		// TODO:GG 修改avatarUrl
		avatarUrl := c.DefaultQuery("avatarUrl", "")
		account.SetAvatarUrl(&avatarUrl)
		// 修改url
		//ok, codeError := service.UpdClientAvatarUrl(u, avatarUrl)
		log.Debug("UpdAccount", log.FAny("avatarUrl", account))
	}

	c.String(http.StatusOK, "account:%s by %s", id, action)
}

func GetAccount(c *gin.Context) {
	id := c.Param("id")

	//firstname := c.DefaultQuery("firstname", "Guest")
	//lastname := c.Query("lastname")

	c.String(http.StatusOK, "account:%s", id)

	//name := c.Param("id")
	//u, err := strconv.ParseUint(name, 10, 64)
	//if err != nil {
	//	c.FString(http.StatusBadRequest, "Invalid ID")
	//	return
	//}
	//client, codeError := service.GetClient(u)
	//if codeError != nil {
	//	c.JSON(http.StatusNotFound, client)
	//}
	//c.JSON(http.StatusOK, client)
}
