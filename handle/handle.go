package handle

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"golocaldownload/common"
	"log"
	"net/http"
	"os"
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
	Name       string  `json:"name"`
	IsDir      bool    `json:"is_dir"`
	Size       float64 `json:"size"`
	SizeUnit   string  `json:"size_unit"`
	ModTime    string  `json:"mod_time"`
	DirPath    string  `json:"dir_path"`
	DirnameKey string  `json:"dirname_key"`
}

func (*Handle) List(ctx *gin.Context) {
	const PathSep = string(os.PathSeparator)
	rootDir := os.Getenv("GLD_download_lib_path")

	dir := rootDir
	dirPath := strings.Trim(ctx.Query("dir_path"), " ")

	if len(dirPath) > 0 {
		dir = dir + dirPath
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
		size, unit := common.FileSizeFormat(uint64(info.Size()))
		list = append(list, FileEntry{
			Name:       paths.Name(),
			IsDir:      paths.IsDir(),
			ModTime:    info.ModTime().Format("2006/01/02 15:04"),
			Size:       common.KeepDecimals(size, 2),
			SizeUnit:   unit,
			DirPath:    dirPath + PathSep + info.Name(),
			DirnameKey: base64.URLEncoding.EncodeToString([]byte(dir + PathSep + paths.Name())),
		})
	}

	ctx.JSON(http.StatusOK, OutEntry{
		RootDir:      rootDir,
		AbsoluteDir:  rootDir + dirPath,
		RelativeDirs: common.StrPathToStrPaths(dirPath, PathSep),
		List:         list,
	})
	return
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
