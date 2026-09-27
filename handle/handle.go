// Package handle 实现下载库的三个 HTTP 接口：目录列表、全局检索、文件下载。
//
// 三个接口都只允许访问下载库根目录之内的内容，所有来自请求的路径都必须经过
// resolve / common.SafeJoin 做越界校验。
package handle

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"golocaldownload/common"
)

// PathSep 是当前操作系统的路径分隔符；接口返回的相对路径都用它拼接。
const PathSep = string(os.PathSeparator)

// maxSearchResults 限制单次检索返回的条目数：下载库可能有几十万个文件，
// 不设上限时一次检索就能把响应体和前端 DOM 一起撑爆。
const maxSearchResults = 500

// Handler 持有下载库根目录，所有接口都只在该目录内工作。
// 根目录改为显式注入，不再通过环境变量 GLD_download_lib_path 在包里隐式传递。
type Handler struct {
	root string
}

// New 创建 Handler，root 为下载库根目录（推荐传绝对路径）。
func New(root string) *Handler {
	return &Handler{root: filepath.Clean(root)}
}

// Root 返回下载库根目录。
func (h *Handler) Root() string { return h.root }

// OutEntry 是目录列表接口的响应体。
type OutEntry struct {
	RootDir      string              `json:"root_dir"`
	AbsoluteDir  string              `json:"absolute_dir"`
	RelativeDirs []map[string]string `json:"relative_dirs"`
	List         []FileEntry         `json:"list"`
}

// FileEntry 是列表 / 检索接口中的单个条目。
type FileEntry struct {
	Name        string  `json:"name"`
	IsDir       bool    `json:"is_dir"`
	Size        float64 `json:"size"`
	SizeUnit    string  `json:"size_unit"`
	ModTime     string  `json:"mod_time"`
	ParentPath  string  `json:"parent_path"`
	Path        string  `json:"path"`
	PathnameKey string  `json:"pathname_key"`
}

// List 列出下载库中指定目录的内容。GET /api/list?path=/a/b
func (h *Handler) List(ctx *gin.Context) {
	dir, err := common.SafeJoin(h.root, ctx.Query("path"))
	if err != nil {
		fail(ctx, http.StatusBadRequest, err)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		fail(ctx, statusOf(err), fmt.Errorf("读取目录失败: %w", err))
		return
	}

	display := displayPath(h.root, dir)
	list := make([]FileEntry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// 单个条目读不到（权限不足、被并发删除等）不该让整个列表失败。
			continue
		}
		list = append(list, h.newEntry(info, dir, display))
	}
	sortEntries(list)

	ctx.JSON(http.StatusOK, OutEntry{
		RootDir:      h.root,
		AbsoluteDir:  dir,
		RelativeDirs: common.StrPathToStrPaths(display, PathSep),
		List:         list,
	})
}

// Download 下载文件。GET /api/download?data=<base64 路径>
func (h *Handler) Download(ctx *gin.Context) {
	decoded, err := base64.URLEncoding.DecodeString(ctx.Query("data"))
	if err != nil {
		fail(ctx, http.StatusBadRequest, errors.New("下载参数格式错误"))
		return
	}

	file, err := h.resolve(string(decoded))
	if err != nil {
		fail(ctx, http.StatusForbidden, err)
		return
	}

	info, err := os.Stat(file)
	if err != nil {
		fail(ctx, statusOf(err), fmt.Errorf("文件不可访问: %w", err))
		return
	}
	if info.IsDir() {
		fail(ctx, http.StatusBadRequest, errors.New("不支持下载目录"))
		return
	}

	// 中文、空格等非 ASCII 文件名必须按 RFC 5987 编码，否则浏览器拿到的是乱码文件名；
	// mime.FormatMediaType 会自动决定用 filename 还是 filename* 形式。
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": info.Name()})
	ctx.Header("Content-Disposition", disposition)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(file)
}

// Search 在下载库中按文件名做全局检索。POST /api/search（表单参数 keyword）
func (h *Handler) Search(ctx *gin.Context) {
	list := make([]FileEntry, 0)
	keyword := strings.TrimSpace(ctx.PostForm("keyword"))
	if keyword == "" {
		ctx.JSON(http.StatusOK, list)
		return
	}

	// WalkDir 不额外 lstat、不跟随符号链接，比 filepath.Walk 更省系统调用。
	err := filepath.WalkDir(h.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if path == h.root {
				return err // 根目录都读不了，直接上报
			}
			return nil // 单个条目不可读时跳过，不影响整体检索
		}
		if path == h.root || !common.FuzzyMatch(entry.Name(), keyword) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		dir := filepath.Dir(path)
		list = append(list, h.newEntry(info, dir, displayPath(h.root, dir)))
		if len(list) >= maxSearchResults {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		log.Printf("检索失败: %v", err)
	}

	ctx.JSON(http.StatusOK, list)
}

// resolve 把客户端传来的路径解析为下载库之内的绝对路径。
//
// 兼容两种写法：相对路径（推荐，如 "/a/b.txt"）与绝对路径（旧版前端会把
// root 前缀一起 base64 编码进去）。两种写法都会校验是否越出 root。
func (h *Handler) resolve(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("路径不能为空")
	}
	if filepath.IsAbs(path) {
		abs := filepath.Clean(path)
		if !common.Inside(h.root, abs) {
			return "", errors.New("拒绝访问下载库之外的路径")
		}
		return abs, nil
	}
	return common.SafeJoin(h.root, path)
}

// newEntry 由文件信息组装一个条目。dir 是该文件所在目录的绝对路径，
// display 是同一目录相对下载库的展示路径。
func (h *Handler) newEntry(info fs.FileInfo, dir, display string) FileEntry {
	size, unit := common.FileSizeFormat(uint64(info.Size()))
	return FileEntry{
		Name:        info.Name(),
		IsDir:       info.IsDir(),
		Size:        common.KeepDecimals(size, 2),
		SizeUnit:    unit,
		ModTime:     info.ModTime().Format("2006/01/02 15:04"),
		ParentPath:  display,
		Path:        display + PathSep + info.Name(),
		PathnameKey: base64.URLEncoding.EncodeToString([]byte(filepath.Join(dir, info.Name()))),
	}
}

// sortEntries 让目录排在文件前面，同类按名称排序（不区分大小写）。
func sortEntries(list []FileEntry) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].IsDir != list[j].IsDir {
			return list[i].IsDir
		}
		return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
	})
}

// displayPath 把绝对路径转成相对下载库的展示路径：根目录返回 ""，子目录返回
// "/a/b" 或 "\a\b"（随系统分隔符）。
func displayPath(root, abs string) string {
	return strings.TrimPrefix(abs, root)
}

// statusOf 把文件系统错误映射成合适的 HTTP 状态码。
func statusOf(err error) int {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return http.StatusNotFound
	case errors.Is(err, fs.ErrPermission):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// fail 统一错误响应。
//
// 旧实现在 handler 里用 log.Panicln 处理错误，而 release 模式创建引擎用的是
// gin.New()（不含 Recovery 中间件），一次「目录不存在」就能让整个进程崩溃。
func fail(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, gin.H{"error": err.Error()})
}
