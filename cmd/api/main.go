package main

import (
	"fmt"
	"katydid-mp-user/api"
	_ "katydid-mp-user/init"
	"katydid-mp-user/pkg/log"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download

// TODO:GG mysql/pgsql/mongoDB/redis
// TODO:GG 加解密
// TODO:GG 异常捕获 sentry
// TODO:GG 监控报警 Prometheus + Grafana
// TODO:GG 链路跟踪
// TODO:GG 超时控制
// TODO:GG 自动熔断 + 限流 + 降载
// TODO:GG etc分布式 + 配置中心
// TODO:GG 代理 + 网关 + 负载均衡
// TODO:GG 服务注册 + 发现 + 治理
// TODO:GG 服务调用 + 网格 + 监控
// TODO:GG 支付stripe
// TODO:GG 定时 cron
// TODO:GG 二维码
// TODO:GG 各种压测(k6)/分析框架

// TODO:GG prometheus + grafana + gatus
// TODO:GG tidb

// TODO:GG go get github.com/gin-contrib/cors
// TODO:GG go get github.com/gin-contrib/sessions

// TODO:GG github.com/swaggo/files v1.0.1
// TODO:GG github.com/swaggo/gin-swagger v1.6.0
// TODO:GG github.com/swaggo/swag v1.16.4
// TODO:GG gorm.io/driver/postgres v1.5.11

// TODO:GG kong +  traefik + Caddy
// TODO:GG gron
// TODO:GG Fluentd
// TODO:GG harness
// TODO:GG Istio
// TODO:GG gods
// TODO:GG go-kit
// TODO:GG ants
// TODO:GG gocron
// TODO:GG nats + watermill + asynq
// TODO:GG sonyflake
// TODO:GG gosms
// TODO:GG gopsutil
// TODO:GG gofakeit
// TODO:GG go-pinyin
// TODO:GG lego
// TODO:GG certmagic
// TODO:GG protobuf
// TODO:GG gjson
// TODO:GG email
// TODO:GG decimal
// TODO:GG cobra
// TODO:GG afero
// TODO:GG go-callvis
// TODO:GG hey + vegeta + Comcast

func main() {
	// 写一个channel，让main挂起来
	log.Debug("这里是debug")
	log.Info("这里是info")
	log.Warn("这里是warn")
	//log.Error("这里是error")

	defer func() {
		err := log.Close()
		if err != nil {
			fmt.Printf("关闭日志失败: %v\n", err) // TODO:GG
		}
	}()

	//client := model.NewClientEmpty()
	//client.Name = ""
	//v := model2.NewValidator(client)
	//err := v.Valid(model2.ValidSceneAll)
	//err := validate.Struct(client)
	//if err != nil {
	//	fmt.Println(err)
	//}

	//go func() {
	//	for i := 0; true; i++ {
	//		//time.Sleep(100 * time.Millisecond)
	//		time.Sleep(1 * time.Second)
	//		fmt.Printf("i: %d\n", i)
	//		log.Debug("这里是debug这里是debug这里是debug这里是debug这里是debug", log.FInt("i", i))
	//		log.Info("这里是info这里是info这里是info这里是info这里是info", log.FInt("i", i))
	//		log.Warn("这里是warn这里是warn这里是warn这里是warn这里是warn", log.FInt("i", i))
	//		//log.Error("这里是error这里是error这里是error这里是error这里是error", log.FInt("i", i))
	//	}
	//}()

	//localize := i18n.LocalizeTry("zh-CN", "hellos", map[string]any{
	//	"Name":  "GG",
	//	"Count": 18,
	//})
	//log.Info("localize -->", log.FString("localize", localize))

	//ch := make(chan int)
	//for {
	//	<-ch
	//}
	api.Run()
}
