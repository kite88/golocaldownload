// Package common 放与业务无关的通用工具：目录准备、路径安全解析、体积格式化。
package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// EnsureDir 确保 dirPath 对应的目录存在（含多级父目录），并返回其绝对路径。
// dirPath 为空时返回空串，交由调用方决定默认目录。
//
// 旧实现用的是 os.Mkdir（只能建一级目录），配置里写 ./a/b 这种路径会直接失败。
func EnsureDir(dirPath string) (string, error) {
	dirPath = strings.TrimSpace(dirPath)
	if dirPath == "" {
		return "", nil
	}
	abs, err := filepath.Abs(dirPath)
	if err != nil {
		return "", fmt.Errorf("解析目录 %s 的绝对路径失败: %w", dirPath, err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", fmt.Errorf("创建目录 %s 失败: %w", abs, err)
	}
	return abs, nil
}

// SafeJoin 把子路径 rel 拼到 base 之下，并保证结果不会逃出 base。
//
// 来自 HTTP 请求的路径必须走这里：旧实现直接做字符串拼接（base + path），
// 攻击者只要传 path=/../../etc/passwd 就能列目录、甚至借助下载接口读取
// 服务端的任意文件（目录穿越）。这里统一做三件事：
//  1. 拒绝含 NUL 的路径；
//  2. 统一分隔符、清掉首部斜杠，让 "/a/b"、"a/b"、"a\b" 等价；
//  3. 用 filepath.Rel 确认 base 到目标的相对路径不以 ".." 开头。
func SafeJoin(base, rel string) (string, error) {
	if strings.ContainsRune(rel, 0) {
		return "", fmt.Errorf("路径包含非法字符")
	}
	rel = filepath.FromSlash(rel)
	rel = strings.TrimLeft(rel, string(filepath.Separator)+`/\`)
	if rel == "" || rel == "." {
		return filepath.Clean(base), nil
	}

	// Join 内部会做 Clean，"/a/../b" 这类写法会被正常归一化。
	abs := filepath.Join(base, rel)
	if !Inside(base, abs) {
		return "", fmt.Errorf("路径越界: %s", rel)
	}
	return abs, nil
}

// Inside 判断 target 是否位于 base 目录之内（含 base 自身）。
// 比较前会把两侧都转成绝对路径并归一化，因此 base 写法不同也能正确判断。
func Inside(base, target string) bool {
	base = normalize(base)
	target = normalize(target)

	diff, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	if diff == "." || diff == "" {
		return true
	}
	return diff != ".." && !strings.HasPrefix(diff, ".."+string(filepath.Separator))
}

func normalize(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return filepath.Clean(path)
}

// FileSizeFormat 把字节数换算成人类可读的数值与单位（如 1.50, "MB"）。
// 数值与单位分开返回，是为了让前端能按自己的方式排版。
func FileSizeFormat(size uint64) (newSize float64, unitS string) {
	newSize = float64(size)
	if size < 1024 {
		return newSize, "Bytes"
	}

	unit := []string{"KB", "MB", "GB", "TB", "PB"}
	var unitIndex int
	for i := 0; i < len(unit); i++ {
		newSize /= 1024
		unitIndex = i
		if newSize < 1024 {
			break
		}
	}
	return newSize, unit[unitIndex]
}

// KeepDecimals 保留 decimals 位小数。
func KeepDecimals(number float64, decimals int) float64 {
	format := "%." + strconv.Itoa(decimals) + "f"
	newNumber, err := strconv.ParseFloat(fmt.Sprintf(format, number), 64)
	if err != nil {
		// 参数只可能来自内部常量，理论上不会失败；真失败了也不值得中断请求。
		return number
	}
	return newNumber
}

// StrPathToStrPaths 把 "/a/b" 这样的路径拆成逐级累加的面包屑：
//
//	[{"" :""}, {"a":"/a"}, {"b":"/a/b"}]
//
// 空键代表根目录（前端据此显示「根目录」）。返回值沿用 []map 结构，
// 是为了与既有前端契约保持兼容。
func StrPathToStrPaths(strPath string, sep string) []map[string]string {
	result := make([]map[string]string, 0)
	var dirPath string
	for _, path := range strings.Split(strPath, sep) {
		if len(path) > 0 {
			dirPath += sep + path
		}
		result = append(result, map[string]string{path: dirPath})
	}
	return result
}

// FuzzyMatch 模糊匹配函数（简单子字符串匹配，忽略大小写）。
func FuzzyMatch(filename, keyword string) bool {
	return strings.Contains(strings.ToLower(filename), strings.ToLower(keyword))
}
