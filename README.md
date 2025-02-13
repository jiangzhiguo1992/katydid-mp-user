# katydid-mp-user
业务中台 (账户+权限+用户+组织+终端+etc)

katydid-mp-user 是一个基于 __golang__ 的 __api__ 项目，用于快速搭建api项目的环境，提供了一些基础的功能，功能如下：

- 1.__版本控制+统计系统__
- 2.__用户认证+权限+统计系统__
- 3.__账号认证+权限+统计系统__
- 4.__公共消息(公告+客服)系统__
- 5.__支付系统__
- 6.__日志系统__

所用到的框架如下：

- 1.__gin__ 用于搭建web服务
- 2.__gorm__ 用于操作数据库 (包括中间件 postgres、redis)
- 3.__viper__ 用于读取本地/分布式配置文件
- 4.__zap__ 用于日志记录 (关联fsnotify)
- 5.__testify__ 用于单元测试
- 6.__swag__ swagger文档(关联files，中间件gin-swagger)
- 7.__i18n__ 国际化(关联text)
- 8.__toml__ toml配置文件解析
- 9.__cobra__ 命令行工具

## 运行项目可执行如下操作

### 1. 首先是构建镜像

#### 1.构建单个镜像 Dockerfile (可以跳过，直接进行第2步)

#### dev:

```shell
    docker build -f deployments/docker/dev/client/Dockerfile -t  katydid_base_api-client_dev .
```

#### pro:

```shell
    docker build -f deployments/docker/prod/client/Dockerfile -t katydid_base_api-client_prod .
```

- -f是指定的Dockerfile文件位置，-t是指定的镜像名称，其中 katydid_base_api-client 替换为自己的docker镜像名称，.是指上下文

#### 2.构建组合镜像 docker-compose.yml

#### dev:

```shell
    docker-compose -f deployments/docker/dev/docker-compose.yml up --build -d
```

#### prod:

```shell
    docker-compose -f deployments/docker/pord/docker-compose.yml up --build -d
```

- -f是指定的docker-compose.yml文件位置，--build强制构建(不使用缓存)，-d后台运行

### 运行swagger文档

```shell
    swag init -g ./cmd/api/main.go -o ./docs/swagger
```

