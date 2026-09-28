// Package test 用进程内路由做集成测试：不占端口、不依赖外部环境，
// 直接验证「HTTP 请求 -> 路由 -> handler -> 磁盘」这条链路。
package test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"golocaldownload/handle"
	"golocaldownload/router"
)

// fixture 是一次测试所需的临时环境。
type fixture struct {
	engine *gin.Engine
	root   string // 下载库根目录
	secret string // 下载库之外的「机密」文件，用于验证目录穿越已被拦住
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	gin.SetMode(gin.TestMode)

	root := t.TempDir()
	outside := t.TempDir()

	writeFile(t, filepath.Join(root, "readme.txt"), "hello")
	writeFile(t, filepath.Join(root, "中文名称.txt"), "中文内容")
	writeFile(t, filepath.Join(root, "sub", "inner.bin"), strings.Repeat("x", 64))
	writeFile(t, filepath.Join(root, "sub", "nested", "deep.txt"), "deep")

	secret := filepath.Join(outside, "secret.txt")
	writeFile(t, secret, "top secret")

	// 模板与静态资源取自真实文件：测试目录的上一层就是项目根目录。
	engine, err := router.R(gin.TestMode, os.DirFS(".."), os.DirFS(".."), handle.New(root, ""))
	if err != nil {
		t.Fatalf("创建路由失败: %v", err)
	}
	return &fixture{engine: engine, root: root, secret: secret}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) do(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	f.engine.ServeHTTP(rec, req)
	return rec
}

func (f *fixture) get(t *testing.T, path string, params url.Values) *httptest.ResponseRecorder {
	t.Helper()
	target := path
	if len(params) > 0 {
		target += "?" + params.Encode()
	}
	return f.do(t, httptest.NewRequest(http.MethodGet, target, nil))
}

func (f *fixture) post(t *testing.T, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return f.do(t, req)
}

func findEntry(t *testing.T, list []handle.FileEntry, name string) handle.FileEntry {
	t.Helper()
	for _, e := range list {
		if e.Name == name {
			return e
		}
	}
	t.Fatalf("列表中没有找到 %q", name)
	return handle.FileEntry{}
}

func TestIndexPage(t *testing.T) {
	f := newFixture(t)

	rec := f.get(t, "/", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / 状态码 = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "本地文件下载服务") {
		t.Error("首页内容不符合预期")
	}
}

func TestStaticAsset(t *testing.T) {
	f := newFixture(t)

	for _, path := range []string{
		"/web/static/icon/material/folder-base.svg",
		"/web/static/icon/material/zip.svg",
		"/favicon.ico",
	} {
		rec := f.get(t, path, nil)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s 状态码 = %d, want 200", path, rec.Code)
		}
		if got := rec.Header().Get("Cache-Control"); !strings.Contains(got, "max-age") {
			t.Errorf("GET %s 缺少 Cache-Control: %q", path, got)
		}
	}
}

func TestListRoot(t *testing.T) {
	f := newFixture(t)

	rec := f.get(t, "/api/list", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, want 200", rec.Code)
	}

	var out handle.OutEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if out.RootDir != f.root || out.AbsoluteDir != f.root {
		t.Errorf("根目录字段不符合预期: root=%q abs=%q, want %q", out.RootDir, out.AbsoluteDir, f.root)
	}
	// 面包屑：根目录只有一项，键为空
	if len(out.RelativeDirs) != 1 || out.RelativeDirs[0][""] != "" {
		t.Errorf("根目录面包屑不符合预期: %v", out.RelativeDirs)
	}
	// 目录排在文件前面
	if len(out.List) == 0 || !out.List[0].IsDir || out.List[0].Name != "sub" {
		t.Errorf("目录应当排在文件之前: %+v", out.List)
	}
	if e := findEntry(t, out.List, "readme.txt"); e.IsDir || e.SizeUnit != "Bytes" || e.Size != 5 {
		t.Errorf("文件条目不符合预期: %+v", e)
	}
	if e := findEntry(t, out.List, "sub"); e.Path == "" || !e.IsDir {
		t.Errorf("目录条目不符合预期: %+v", e)
	}
}

func TestListSubDir(t *testing.T) {
	f := newFixture(t)

	rec := f.get(t, "/api/list", url.Values{"path": {"/sub"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, want 200", rec.Code)
	}

	var out handle.OutEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.AbsoluteDir != filepath.Join(f.root, "sub") {
		t.Errorf("AbsoluteDir = %q, want %q", out.AbsoluteDir, filepath.Join(f.root, "sub"))
	}
	e := findEntry(t, out.List, "inner.bin")
	if e.ParentPath != string(os.PathSeparator)+"sub" {
		t.Errorf("ParentPath = %q, want %q", e.ParentPath, string(os.PathSeparator)+"sub")
	}
}

func TestListMissingDir(t *testing.T) {
	f := newFixture(t)

	if rec := f.get(t, "/api/list", url.Values{"path": {"/not-exist"}}); rec.Code != http.StatusNotFound {
		t.Errorf("不存在的目录状态码 = %d, want 404", rec.Code)
	}
}

// TestListRejectsTraversal 是本次安全修复的回归用例：
// 旧实现用字符串拼接 root + path，传 /../ 就能列出下载库之外的目录。
func TestListRejectsTraversal(t *testing.T) {
	f := newFixture(t)

	paths := []string{"/..", "/../..", "/../../", "/sub/../../../", ".."}
	for _, p := range paths {
		rec := f.get(t, "/api/list", url.Values{"path": {p}})
		if rec.Code != http.StatusBadRequest {
			t.Errorf("越界路径 %q 状态码 = %d, want 400", p, rec.Code)
		}
		if body := rec.Body.String(); strings.Contains(body, "secret") {
			t.Errorf("越界路径 %q 泄露了下载库之外的内容: %s", p, body)
		}
	}
}

func TestDownloadFile(t *testing.T) {
	f := newFixture(t)

	rec := f.get(t, "/api/download", url.Values{"data": {encode(f.root, "readme.txt")}})
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "hello" {
		t.Errorf("文件内容 = %q, want hello", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != "attachment; filename=readme.txt" {
		t.Errorf("Content-Disposition = %q", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Errorf("Content-Type = %q", got)
	}
}

// TestDownloadNonASCIIFilename 验证中文文件名使用 RFC 5987 编码，
// 否则浏览器会把 Content-Disposition 里的原始字节显示成乱码。
func TestDownloadNonASCIIFilename(t *testing.T) {
	f := newFixture(t)

	rec := f.get(t, "/api/download", url.Values{"data": {encode(f.root, "中文名称.txt")}})
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, want 200", rec.Code)
	}
	got := strings.ToLower(rec.Header().Get("Content-Disposition"))
	if !strings.Contains(got, "utf-8''") {
		t.Errorf("非 ASCII 文件名未按 RFC 5987 编码: %q", got)
	}
}

func TestDownloadMissingAndDir(t *testing.T) {
	f := newFixture(t)

	if rec := f.get(t, "/api/download", url.Values{"data": {encode(f.root, "nope.txt")}}); rec.Code != http.StatusNotFound {
		t.Errorf("下载不存在的文件状态码 = %d, want 404", rec.Code)
	}
	if rec := f.get(t, "/api/download", url.Values{"data": {encode(f.root, "sub")}}); rec.Code != http.StatusBadRequest {
		t.Errorf("下载目录状态码 = %d, want 400", rec.Code)
	}
	if rec := f.get(t, "/api/download", url.Values{"data": {"不是-base64"}}); rec.Code != http.StatusBadRequest {
		t.Errorf("非法 data 参数状态码 = %d, want 400", rec.Code)
	}
}

// TestDownloadRejectsOutside 覆盖「绝对路径 + 相对路径」两种穿越方式。
func TestDownloadRejectsOutside(t *testing.T) {
	f := newFixture(t)

	absolute := base64.URLEncoding.EncodeToString([]byte(f.secret))
	relative := base64.URLEncoding.EncodeToString([]byte("../" + filepath.Base(f.secret)))

	for name, data := range map[string]string{"绝对路径": absolute, "相对路径": relative} {
		rec := f.get(t, "/api/download", url.Values{"data": {data}})
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s 越界下载状态码 = %d, want 403", name, rec.Code)
		}
		if body := rec.Body.String(); strings.Contains(body, "top secret") {
			t.Errorf("%s 越界下载泄露了下载库之外的文件", name)
		}
	}
}

func TestSearch(t *testing.T) {
	f := newFixture(t)

	rec := f.post(t, "/api/search", url.Values{"keyword": {"inner"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, want 200", rec.Code)
	}
	var list []handle.FileEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "inner.bin" {
		t.Fatalf("检索结果不符合预期: %+v", list)
	}
	if list[0].ParentPath != string(os.PathSeparator)+"sub" {
		t.Errorf("ParentPath = %q", list[0].ParentPath)
	}

	// 命中目录名
	rec = f.post(t, "/api/search", url.Values{"keyword": {"NESTED"}})
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || !list[0].IsDir || list[0].Name != "nested" {
		t.Fatalf("目录检索结果不符合预期: %+v", list)
	}

	// 空关键字与无命中都应返回 200 + 空数组（而不是 null）
	for _, keyword := range []string{"", "   ", "绝不会命中的关键字"} {
		rec = f.post(t, "/api/search", url.Values{"keyword": {keyword}})
		if rec.Code != http.StatusOK {
			t.Fatalf("关键字 %q 状态码 = %d, want 200", keyword, rec.Code)
		}
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Errorf("关键字 %q 响应 = %s, want []", keyword, got)
		}
	}
}

func TestNoRoute(t *testing.T) {
	f := newFixture(t)

	if rec := f.get(t, "/api/nope", nil); rec.Code != http.StatusNotFound {
		t.Errorf("未知接口状态码 = %d, want 404", rec.Code)
	}
	if rec := f.get(t, "/nope", nil); rec.Code != http.StatusNotFound {
		t.Errorf("未知页面状态码 = %d, want 404", rec.Code)
	}
}

// encode 复现前端把「绝对路径」base64 后放进 data 参数的写法。
func encode(root, name string) string {
	return base64.URLEncoding.EncodeToString([]byte(filepath.Join(root, name)))
}
