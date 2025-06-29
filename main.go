package main

import (
	"embed"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golocaldownload/common"
	"golocaldownload/config"
	"golocaldownload/router"
	"log"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

var (
	//go:embed web/view/*
	viewFS embed.FS
	//go:embed web/static/*
	staticFs embed.FS

	downloadLibPath = config.GetValue("download_lib_path")
	protocol        = config.GetValue("server.protocol")
	httpPort        = config.GetValue("server.http_port")
)

func main() {
	// 设置环境模式
	var envMode = config.GetValue("env_mode")
	gin.SetMode(envMode)

	defer exceptionalHandling(envMode)

	fmt.Println(time.Now().Format(time.DateTime), envMode)

	downloadLibPathInit()

	outLocalIPs(protocol, httpPort)

	r := router.R(envMode, viewFS, staticFs)

	if err := r.Run(":" + httpPort); err != nil {
		panic(errors.New("程序启动异常：" + err.Error()))
	}
}

func downloadLibPathInit() {
	dir, err := common.GetDirPath(downloadLibPath)
	if err != nil {
		panic(err)
	}
	if len(dir) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		dir = wd
	}
	if err = os.Setenv("GLD_download_lib_path", dir); err != nil {
		panic(err)
	}
	fmt.Println("下载库目录:", dir)
}

func exceptionalHandling(envMode string) {
	if err := recover(); err != nil {
		if envMode != gin.ReleaseMode {
			log.Panicf("发生异常#%v", strings.Trim(string(debug.Stack()), "\n"))
			return
		}
		log.Println("程序发生异常#", err)
		return
	}
}

func outLocalIPs(protocol, httpPort string) {
	addrArr, err := net.InterfaceAddrs()
	if err != nil {
		log.Panicf("net.InterfaceAddrs %v", err)
		return
	}
	for _, addr := range addrArr {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				fmt.Printf("%s://%s:%s\n", protocol, ipNet.IP.String(), httpPort)
			}
		}
	}
}
