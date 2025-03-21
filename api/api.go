package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"katydid-mp-user/api/app"
	"katydid-mp-user/configs"
	"katydid-mp-user/pkg/log"
	"katydid-mp-user/pkg/middleware"
	"net/http"
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

	// TODO:GG swagger

	// panic (捕捉全局)
	engine.Use(gin.Recovery())

	// 日志 // TODO:GG conf自定义
	engine.Use(middleware.ZapLoggerWithConfig(
		middleware.DefaultLoggerConfig(),
	))

	// 国际化 // TODO:GG conf自定义
	engine.Use(middleware.Language())

	// 限流 // TODO:GG conf自定义 + 整体限流(代理做)?
	//engine.Use(middleware.RateLimiter(1000, time.Minute))
	engine.Use(middleware.NewLimiterWithOptions(middleware.LimiterOptions{
		Code:         http.StatusTooManyRequests, // toast
		Message:      "IP请求过于频繁，请稍后再试",
		WhitelistIPs: []string{},
		KeyFunc:      middleware.IPKeyFunc, // ip限流
		DefLimit:     1000,
		DefDuration:  time.Minute,
	}, middleware.LimitRule{
		// TODO:GG ip限流规则 (发送验证码?)
	}, middleware.LimitRule{
		// TODO:GG ip限流规则 (和钱相关的api?)
	}).Middleware())
	engine.Use(middleware.NewLimiterWithOptions(middleware.LimiterOptions{
		Code:         http.StatusTooManyRequests, // toast
		Message:      "账号请求过于频繁，请稍后再试",
		WhitelistIPs: []string{},
		KeyFunc:      middleware.AccountKeyFunc, // 账号限流
		DefLimit:     1000,
		DefDuration:  time.Minute,
	}, middleware.LimitRule{
		// TODO:GG acc限流规则 (发送验证码?)
	}, middleware.LimitRule{
		// TODO:GG acc限流规则 (和钱相关的api?)
	}).Middleware())

	// 跨域 // TODO:GG conf自定义
	engine.Use(middleware.Cors(
		middleware.DefaultCorsOptions(),
	))

	// 认证 // TODO:GG conf自定义
	jwtSecret := ""
	engine.Use(middleware.Auth(middleware.DefaultAuthConfig(jwtSecret, []string{
		"/api/v1/auth",   // 认证
		"/api/v1/verify", // 验证码
	})))

	// TODO:GG 权限 conf自定义
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

	// 缓存 // TODO:GG conf自定义
	engine.Use(middleware.Cache(
		middleware.DefaultCacheConfig(map[string]*time.Duration{
			"/api/v1/article": nil, // 文章
		}),
	))

	// TODO:GG 静态文件  // TODO:GG conf自定义
	//engine.StaticFile()
	engine.GET("/local/file", func(c *gin.Context) {
		// TODO:GG admin权限->查看配置
		c.File("./configs/api/init.toml")
	})

	// api路由
	router := engine.Group("api/v1")
	app.RouterRegister(router)

	host := "" // TODO:GG api.katydid.com
	port := configs.Get().Server.ApiHttpsPort
	err := engine.Run(fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatal("gin run panic", log.FError(err))
	}
	return engine
}
