package app

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	accountHandler "katydid-mp-user/internal/api/account/handler"
	clientHandler "katydid-mp-user/internal/api/client/handler"
	"katydid-mp-user/internal/api/perm/handler"
	BHandler "katydid-mp-user/internal/pkg/handler"
	"katydid-mp-user/pkg/middleware"
)

type FormAccount struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var fun = func(base *BHandler.Base, handler1 func()) func(context *gin.Context) {
	return func(context *gin.Context) {
		base.SetGCtx(context)
		handler1()
	}
}

func RouterRegister(r *gin.RouterGroup, authManager *middleware.AuthManager) {
	v1 := r.Group("v1")
	//
	//e := handler.ValidPwd()
	//if e != nil {
	//	log.Error("ValidPwd", log.Error(e))
	//}

	// TODO:GG 抽到configs里面吗？
	//swagger.SwaggerInfo.Title = "Swagger Example API"
	//swagger.SwaggerInfo.Description = "This is a sample server Petstore server."
	//swagger.SwaggerInfo.Version = "1.0"
	//swagger.SwaggerInfo.Host = "petstore.swagger.io"
	//swagger.SwaggerInfo.BasePath = "/api/v1"
	//swagger.SwaggerInfo.Schemes = []string{"http", "https"}

	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//ginSwagger.WrapHandler()

	// 注册验证器
	//if err := validation.Register(); err != nil {
	//	log.Fatal("Failed to register validators:", log.Err(err))
	//}

	// TODO:GG 加一个所有的handler的before和after？在哪来一个handler的多态，控制某类handler的before和after
	// TODO:GG 添加filters，如auth，adminkey，cache，permission，localize，cors(web)/devices(mobile)等
	// TODO:GG auth思路，client持有sessionId/token/cookies，后端持有session/token，且是jwt格式，方便做过期，然后根据sessionId来检索后端的session/token
	// TODO:GG 可以做多级缓存，http-cache(自带的session存储)，redis-cache，mysql-cache
	// TODO:GG 关键操作还得让+verify(email/phone)

	// orga
	oh := handler.NewOrganization(nil)
	v1.Use(func(context *gin.Context) {

	})
	{
		team := v1.Group("organization")
		team.POST("", fun(oh.Base, oh.PostTeam))
		//team.POST("", ctx.Post)
	}

	// account
	account := v1.Group("account")
	{
		account.POST("", accountHandler.PostAccount)
		account.PUT(":id/*action", accountHandler.PutAccount)
		account.GET(":id", accountHandler.GetAccount)
	}

	//// 登录接口 - 不需要认证
	//r.POST("/login", func(c *gin.Context) {
	//	// 验证用户凭据...
	//	token, err := authManager.GenerateToken("123", "testuser", []string{"user"})
	//	if err != nil {
	//		c.JSON(500, gin.H{"error": err.Error()})
	//		return
	//	}
	//	c.JSON(200, gin.H{"token": token})
	//})
	//
	//// 需要admin角色的接口
	//admin := r.Group("/admin")
	//admin.Use(authManager.RoleCheck("admin"))
	//{
	//	admin.GET("/dashboard", func(c *gin.Context) {
	//		c.JSON(200, gin.H{"msg": "管理员面板数据"})
	//	})
	//}
	//
	//// 登出
	//r.POST("/logout", func(c *gin.Context) {
	//	authManager.InvalidateToken(c)
	//	c.JSON(200, gin.H{"msg": "登出成功"})
	//})

	// account.auth
	v1.Group("account/auth")
	{
	}

	// account.permission
	v1.Group("account/permission")
	{
	}

	// client
	client := v1.Group("client")
	{
		client.POST("", clientHandler.PostClient)
	}

	// user
	v1.Group("user")
	{
	}

	// stats TODO:GG 要放这里吗?

}

//func RegisterClient(r *gin.RouterGroup) {
//	// team
//	r = r.Group("team")
//	{
//		r.GET(":id", handler.GetTeam)
//		r.POST("", handler.PostTeam)
//	}
//
//	// client
//	r = r.Group("client")
//	{
//		r.GET(":id", handler.GetClient)
//		r.POST("", handler.AddClient)
//		r.DELETE(":id", handler.DelClient)
//		//r.GET(":id", handler.GetClient)
//	}
//}
