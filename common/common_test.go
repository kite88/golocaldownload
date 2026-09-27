package common

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestSafeJoinKeepsInsideRoot(t *testing.T) {
	root := t.TempDir()

	for _, rel := range []string{"", "/", "a", "/a", "a/b", "/a/b", "/a/../b"} {
		got, err := SafeJoin(root, rel)
		if err != nil {
			t.Fatalf("SafeJoin(%q) 返回错误: %v", rel, err)
		}
		if !Inside(root, got) {
			t.Fatalf("SafeJoin(%q) = %q 落在了根目录之外", rel, got)
		}
	}

	got, err := SafeJoin(root, "/a/../b")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "b"); got != want {
		t.Fatalf("SafeJoin 归一化结果错误: got %q, want %q", got, want)
	}
}

func TestSafeJoinRejectsEscape(t *testing.T) {
	root := t.TempDir()

	rels := []string{"..", "../..", "/../x", "a/../../x", "/a/../../../../../x", "a/../../../x"}
	if runtime.GOOS == "windows" {
		rels = append(rels, `..\..\x`, `a\..\..\..\x`)
	}
	for _, rel := range rels {
		if _, err := SafeJoin(root, rel); err == nil {
			t.Errorf("SafeJoin(%q) 未被拒绝，存在目录穿越风险", rel)
		}
	}
}

func TestInside(t *testing.T) {
	root := filepath.Join(t.TempDir(), "lib")

	cases := []struct {
		target string
		want   bool
	}{
		{root, true},
		{filepath.Join(root, "a"), true},
		{filepath.Join(root, "a", "b.txt"), true},
		{filepath.Join(root, "..", "sibling"), false},
		{filepath.Dir(root), false},
	}
	for _, c := range cases {
		if got := Inside(root, c.target); got != c.want {
			t.Errorf("Inside(%q, %q) = %v, want %v", root, c.target, got, c.want)
		}
	}
}

func TestFileSizeFormat(t *testing.T) {
	cases := []struct {
		size uint64
		want float64
		unit string
	}{
		{0, 0, "Bytes"},
		{1023, 1023, "Bytes"},
		{1024, 1, "KB"},
		{1536, 1.5, "KB"},
		{1 << 20, 1, "MB"},
		{1 << 30, 1, "GB"},
		{1 << 40, 1, "TB"},
		{1 << 50, 1, "PB"},
		{1 << 60, 1024, "PB"}, // 超过 PB 后继续按 PB 计
	}
	for _, c := range cases {
		size, unit := FileSizeFormat(c.size)
		if size != c.want || unit != c.unit {
			t.Errorf("FileSizeFormat(%d) = %v %s, want %v %s", c.size, size, unit, c.want, c.unit)
		}
	}
}

func TestKeepDecimals(t *testing.T) {
	if got := KeepDecimals(1.239, 2); got != 1.24 {
		t.Errorf("KeepDecimals(1.239, 2) = %v, want 1.24", got)
	}
	if got := KeepDecimals(2, 0); got != 2 {
		t.Errorf("KeepDecimals(2, 0) = %v, want 2", got)
	}
}

func TestStrPathToStrPaths(t *testing.T) {
	got := StrPathToStrPaths("/a/b", "/")
	want := []map[string]string{{"": ""}, {"a": "/a"}, {"b": "/a/b"}}
	if len(got) != len(want) {
		t.Fatalf("长度不一致: got %v, want %v", got, want)
	}
	for i := range want {
		for k, v := range want[i] {
			if got[i][k] != v {
				t.Errorf("第 %d 项 = %v, want %v", i, got[i], want[i])
			}
		}
	}

	if got := StrPathToStrPaths("", "/"); len(got) != 1 || got[0][""] != "" {
		t.Errorf("空路径应当只返回根目录面包屑，got %v", got)
	}
}

func TestFuzzyMatch(t *testing.T) {
	if !FuzzyMatch("Report.PDF", "report") {
		t.Error("匹配应当忽略大小写")
	}
	if FuzzyMatch("Report.PDF", "doc") {
		t.Error("不应命中不相关的关键字")
	}
}
