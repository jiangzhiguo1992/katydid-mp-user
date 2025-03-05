package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"katydid-mp-user/api/api"
	"katydid-mp-user/configs"
	"katydid-mp-user/pkg/log"
	"katydid-mp-user/pkg/middleware"
	"time"
)

func Run() *gin.Engine {
	debug := configs.Get().IsDebug()
	test := configs.Get().IsTest()
	prod := configs.Get().IsProd()
	if debug {
		gin.SetMode(gin.DebugMode)
	} else if test {
		gin.SetMode(gin.TestMode)
	} else if prod {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// TODO:GG 静态文件
	// TODO:GG swagger

	// 捕捉panic
	engine.Use(gin.Recovery())

	// 日志 // TODO:GG conf自定义
	engine.Use(middleware.ZapLoggerWithConfig(
		middleware.DefaultLoggerConfig(),
	)) // gin.Logger() + gin.ForceConsoleColor()

	// 国际化 // TODO:GG conf自定义
	engine.Use(middleware.Language())

	// 跨域 // TODO:GG conf自定义
	engine.Use(middleware.Cors(
		middleware.DefaultCorsOptions(),
	))

	// 使用认证中间件 // TODO:GG conf自定义
	authConfig := middleware.DefaultAuthConfig("sasasa", debug)
	authConfig.TokenExpiration = time.Hour * 24
	authManager := middleware.NewAuthManager(authConfig)
	engine.Use(authManager.Auth())

	// 限流 // TODO:GG conf自定义
	limiter := middleware.NewLimiter(1_000, time.Minute)
	limiter.WithOptions(middleware.LimiterOptions{
		Code:         1, // toast
		Message:      "请求过于频繁，请稍后再试",
		WhitelistIPs: []string{},
		KeyFunc:      middleware.UserKeyFunc,
		LogFunc: func(msg string, args ...any) {
			log.Info(msg, log.FString("args", fmt.Sprintf("%v", args)))
		},
	})
	limiter.AddRule(middleware.LimitRule{})
	engine.Use(limiter.Middleware())

	// 静态文件  // TODO:GG conf自定义
	//engine.StaticFile()
	engine.GET("/local/file", func(c *gin.Context) {
		// TODO:GG admin权限->查看配置
		c.File("./configs/api/init.toml")
	})

	// api路由
	router := engine.Group("api")
	api.RouterRegister(router, authManager)

	//registerRoutesToCasbin(engine)
	//// 使用Casbin鉴权中间件
	//engine.Use(CasbinMiddleware())
	// 设置示例路由
	engine.GET("/data1", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "允许访问/data1"})
	})
	engine.GET("/data2", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "允许访问/data2"})
	})

	port := configs.Get().Server.ApiHttpsPort
	err := engine.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("gin run panic", log.FError(err))
	}
	return engine
}
