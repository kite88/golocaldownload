// Command golocaldownload 是一个把本机目录当「下载库」对外提供 HTTP 访问的小服务：
// 浏览目录、全局检索文件名、下载文件，页面与静态资源全部内嵌在二进制里。
//
// 用法示例：
//
//	golocaldownload                  # 使用内嵌的 release 配置
//	golocaldownload -config env.ini  # 指定外部配置文件
//	golocaldownload -version         # 打印版本号
package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"golocaldownload/common"
	"golocaldownload/config"
	"golocaldownload/handle"
	"golocaldownload/router"
)

// version 由构建脚本用 -ldflags "-X main.version=..." 注入，源码运行时为 dev。
var version = "dev"

var (
	//go:embed web/view/*
	viewFS embed.FS
	//go:embed web/static/*
	staticFS embed.FS
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("程序异常退出: %v", err)
	}
}

func run() error {
	configPath := flag.String("config", "", "外部配置文件路径（ini）；留空时依次查找 ./env.ini、./config/env.ini 与内嵌模板")
	showVersion := flag.Bool("version", false, "打印版本号后退出")
	flag.Parse()

	if *showVersion {
		fmt.Printf("golocaldownload %s\n", version)
		return nil
	}

	if err := config.Init(*configPath); err != nil {
		return err
	}
	envMode := config.EnvMode()
	gin.SetMode(envMode)

	fmt.Printf("%s golocaldownload %s（%s）\n", time.Now().Format(time.DateTime), version, envMode)
	fmt.Println("配置来源:", config.Source())

	libDir, err := prepareDownloadLib()
	if err != nil {
		return err
	}
	fmt.Println("下载库目录:", libDir)

	engine, err := router.R(envMode, viewFS, staticFS, handle.New(libDir))
	if err != nil {
		return err
	}

	port := config.GetString("server.http_port", "9801")
	printAccessURLs(config.GetString("server.protocol", "http"), port)
	return serve(engine, port)
}

// prepareDownloadLib 准备下载库目录并返回其绝对路径。
// 配置留空时退回程序当前工作目录，与旧版行为一致。
func prepareDownloadLib() (string, error) {
	dir, err := common.EnsureDir(config.GetValue("download_lib_path"))
	if err != nil {
		return "", err
	}
	if dir != "" {
		return dir, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	return wd, nil
}

// serve 启动 HTTP 服务，并在收到 Ctrl+C / SIGTERM 时优雅退出。
func serve(engine *gin.Engine, port string) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: engine,
		// 只限制读请求头：下载大文件时写耗时不设上限，避免长传输被强行掐断。
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("HTTP 服务启动失败: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		fmt.Println("收到退出信号，正在关闭服务...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("关闭 HTTP 服务失败: %w", err)
	}
	fmt.Println("服务已停止")
	return nil
}

// printAccessURLs 打印可访问的地址，方便容器/局域网部署时直接复制。
func printAccessURLs(protocol, port string) {
	fmt.Printf("本机访问: %s://127.0.0.1:%s\n", protocol, port)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("获取本机网络地址失败: %v", err)
		return
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}
		fmt.Printf("局域网访问: %s://%s:%s\n", protocol, ipNet.IP, port)
	}
}
