# 构建阶段
FROM golang:1.23-alpine AS builder

# 国内构建加速；需要走官方代理时用 --build-arg GOPROXY_URL=https://proxy.golang.org,direct 覆盖
ARG GOPROXY_URL=https://goproxy.cn,direct
# 版本号会写进二进制的 -version 输出
ARG GLD_VERSION=dev

ENV GO111MODULE=on \
    GOPROXY=${GOPROXY_URL} \
    CGO_ENABLED=0

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags "-s -w -X main.version=${GLD_VERSION}" -o /golocaldownload .

# 运行阶段：只要一个静态二进制，页面模板与静态资源已内嵌
FROM alpine:latest

# 文件修改时间按本地时区展示，alpine 默认没有时区数据
RUN apk add --no-cache tzdata

WORKDIR /root/
COPY --from=builder /golocaldownload ./

# 内嵌默认配置里的 download_lib_path 是相对路径 download_lib，
# 因此下载库落在 /root/download_lib，与 README 中的挂载示例一致
EXPOSE 9801
CMD ["./golocaldownload"]
