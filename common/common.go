package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetDirPath(dirPath string) (string, error) {
	if dirPath == "" {
		return "", nil
	}
	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 目录不存在，创建目录
		err := os.Mkdir(dirPath, 0755) // 权限 0755：rwxr-xr-x
		if err != nil {
			log.Printf("创建目录失败: %v\n", err)
			return "", err
		}
		fmt.Println("目录创建成功:", dirPath)
		return dirPath, nil
	} else if err != nil {
		// 其他错误（如权限不足）
		log.Printf("检查目录失败: %v\n", err)
		return "", err
	}
	return dirPath, nil
}

func GetDirAllFilePaths(dirname string) ([]os.DirEntry, error) {

	infos, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	return infos, err
}

// FileSizeFormat size bytes
func FileSizeFormat(size uint64) (newSize float64, unitS string) {
	newSize = float64(size)

	if size < 1024 {
		return newSize, "Bytes"
	}

	var unit = []string{"KB", "MB", "GB", "TB", "PB"}

	var unitIndex int

	for i := 0; i < len(unit); i++ {
		newSize = newSize / 1024
		unitIndex = i

		if newSize < 1024 {
			break
		}
	}

	return newSize, unit[unitIndex]
}

func KeepDecimals(number float64, decimals int) float64 {
	format := "%." + strconv.Itoa(decimals) + "f"
	newNumber, err := strconv.ParseFloat(fmt.Sprintf(format, number), 64)
	if err != nil {
		// do nothing
		log.Printf("keepDecimals err %v\n", err)
	}
	return newNumber
}

func StrPathToStrPaths(strPath string, sep string) []map[string]string {
	var result = make([]map[string]string, 0)

	strPaths := strings.Split(strPath, sep)
	var dirPath string
	for _, path := range strPaths {
		if len(path) > 0 {
			dirPath += sep + path
		}
		result = append(result, map[string]string{path: dirPath})
	}
	return result
}

func FindModuleRoot(dir string) (string, error) {
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return dir, nil
}
