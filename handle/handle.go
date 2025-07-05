package handle

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"golocaldownload/common"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var Ins = new(Handle)

type Handle struct{}

type OutEntry struct {
	RootDir      string              `json:"root_dir"`
	AbsoluteDir  string              `json:"absolute_dir"`
	RelativeDirs []map[string]string `json:"relative_dirs"`
	List         []FileEntry         `json:"list"`
}

type FileEntry = struct {
	Name        string  `json:"name"`
	IsDir       bool    `json:"is_dir"`
	Size        float64 `json:"size"`
	SizeUnit    string  `json:"size_unit"`
	ModTime     string  `json:"mod_time"`
	ParentPath  string  `json:"parent_path"`
	Path        string  `json:"path"`
	PathnameKey string  `json:"pathname_key"`
}

const PathSep = string(os.PathSeparator)

func (*Handle) List(ctx *gin.Context) {
	var rootDir = os.Getenv("GLD_download_lib_path")

	dir := rootDir

	path := strings.Trim(ctx.Query("path"), " ")
	if len(path) > 0 {
		dir = rootDir + path
	}

	pathRes, err := common.GetDirAllFilePaths(dir)
	if err != nil {
		log.Panicln(err)
		return
	}

	var list = make([]FileEntry, 0)

	for _, paths := range pathRes {
		info, err := paths.Info()
		if err != nil {
			continue
		}
		list = append(list, formatLFileInfo(info, dir, rootDir))
	}

	ctx.JSON(http.StatusOK, OutEntry{
		RootDir:      rootDir,
		AbsoluteDir:  rootDir + path,
		RelativeDirs: common.StrPathToStrPaths(path, PathSep),
		List:         list,
	})
	return
}

func formatLFileInfo(info fs.FileInfo, path string, rootDir string) FileEntry {
	size, unit := common.FileSizeFormat(uint64(info.Size()))
	parentPath := strings.TrimPrefix(path, rootDir)
	finalPath := parentPath + PathSep + info.Name()
	return FileEntry{
		Name:        info.Name(),
		IsDir:       info.IsDir(),
		ModTime:     info.ModTime().Format("2006/01/02 15:04"),
		Size:        common.KeepDecimals(size, 2),
		SizeUnit:    unit,
		ParentPath:  parentPath,
		Path:        finalPath,
		PathnameKey: base64.URLEncoding.EncodeToString([]byte(rootDir + finalPath)),
	}
}

func (*Handle) Download(ctx *gin.Context) {

	data := ctx.Query("data")
	decodeString, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		log.Panicln(err)
		return
	}
	file := string(decodeString)

	pathArr := strings.Split(file, string(os.PathSeparator))
	filename := pathArr[len(pathArr)-1]
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Type", "application/octet-stream")

	// 要下载的文件路径
	filePath := file

	// 发送文件给客户端
	ctx.File(filePath)
	return
}

// Search 遍历目录并搜索文件
func (*Handle) Search(ctx *gin.Context) {
	rootDir := os.Getenv("GLD_download_lib_path")

	var list = make([]FileEntry, 0)
	keyword := strings.Trim(ctx.PostForm("keyword"), " ")
	if len(keyword) == 0 {
		ctx.JSON(http.StatusOK, list)
		return
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == rootDir {
			return nil
		}
		// 如果是文件且匹配关键字
		path = strings.TrimSuffix(path, PathSep+info.Name())
		pathStr := ""
		for i, s := range strings.Split(path, PathSep) {
			if i == 0 {
				continue
			}
			pathStr += PathSep + s
		}

		if common.FuzzyMatch(info.Name(), keyword) {
			list = append(list, formatLFileInfo(info, pathStr, rootDir))
		}
		return nil
	})

	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusOK, list)
		return
	}

	ctx.JSON(http.StatusOK, list)
	return
}
