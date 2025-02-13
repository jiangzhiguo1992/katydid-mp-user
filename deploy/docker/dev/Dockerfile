# 使用官方的 Golang 镜像来创建构建工件，这个镜像包含了 Go 语言的编译器，以及一些构建 Go 应用程序所需的工具
# 这里的go-version需要个mod里的version一致
FROM golang:1.23.4 AS builder
LABEL author="jiang"
#ENV GO111MODULE=on \
#    CGO_ENABLED=0 \
#    GOOS=linux \
#    GOARCH=amd64

# 设置容器内的当前工作目录，类似cd命令
WORKDIR /app

# 复制 go.mod 和 go.sum 文件到工作目录(/app)，下载依赖所需
COPY go.mod go.sum ./

# 下载所有依赖项并构建 Go 应用程序
# 将多个命令合并到一个 RUN 指令中，减少镜像层的数量，但是不能利用缓存
#RUN go mod download && \
#    go build -o main ./cmd \

# 下载所有依赖项。如果 go.mod 和 go.sum 文件没有变化，依赖项将被缓存，不会重新下载
# 这里没有和go-build合并是为了缓存，避免每次构建都下载
RUN go mod download

# 复制当前目录下的所有源代码到工作目录(/app)，类似cp命令
#COPY cmd ./cmd #(其实这个也行，只复制main相关文件进去)
COPY . .

# 构建 Go 应用程序，cmd目录下的man函数，构建二进制文件(-o)名称为main，构建完成后，main文件会在docker的/app目录下
RUN go build -o main ./cmd

# 本地模拟ubuntu24.04，后续的操作都是在ubuntu镜像上进行的
FROM ubuntu:24.04

# 从前一个阶段复制预构建的二进制文件到新的镜像中，这样就不会包含构建工具和源代码，使镜像更加轻量
COPY --from=builder /app/main /app/main

# 向外界暴露端口 8080，这样外部网络可以访问应用程序
EXPOSE 8080

# 声明挂载点，容器内的 /app 目录将作为挂载点
#VOLUME ["/app"]

# 设置容器启动时运行的命令，类似于docker run image_name，这里是运行/app/main
ENTRYPOINT ["/app/main"]

# 可选：添加健康检查，检查应用程序是否正常运行，如果检查失败，容器将退出并容器，达到自愈的效果
# interval=30s 每隔30秒检查一次
# timeout=10s 检查超时时间为10秒
# start-period=5s 容器启动后5秒后开始检查#
# retries=3 检查失败后重试3次
# 这里是检查应用程序的 /health 端点，如果返回状态码不是200，容器就会退出，这样k8s就会重启容器
#HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 CMD curl -f http://localhost:8080/health || exit 1 CMD curl -f http://localhost:8080/health || exit 1| exit 1
