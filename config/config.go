package config

import (
	"embed"
	"gopkg.in/ini.v1"
	"strings"
)

//go:embed env.ini
var envConfigFile embed.FS

var conf *ini.File

func init() {
	Init()
}

// Init 初始化-加载配置文件
func Init() {

	// 读取嵌入的INI文件
	data, err := envConfigFile.ReadFile("env.ini")
	if err != nil {
		panic(err)
	}

	// 解析INI文件
	cfg, err := ini.Load(data)
	if err != nil {
		panic(err)
	}

	conf = cfg
}

// GetValue 使用示例 1、env_mode 2、server.protocol
func GetValue(key string) string {
	result := strings.Split(key, ".")
	if len(result) > 1 {
		return conf.Section(result[0]).Key(result[len(result)-1]).Value()
	}
	return conf.Section("").Key(key).Value()
}
