# 构建 Go 二进制文件
FROM golang:1.23-alpine AS golocaldownload-builder
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN mv -f config/env.ini.release config/env.ini
RUN CGO_ENABLED=0 GOOS=linux go build -o /main-srv

# 生成最小化生产镜像
FROM alpine:latest
WORKDIR /root/
COPY --from=golocaldownload-builder /main-srv ./
EXPOSE 9801
CMD ["./main-srv"]