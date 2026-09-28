package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// restore 把全局配置还原成自动发现的默认值，避免用例之间互相污染。
func restore(t *testing.T) {
	t.Helper()
	// 屏蔽外部环境：本机若导出过 GLD_CONFIG，会让「自动发现」类用例读错文件。
	t.Setenv(PathKey, "")
	t.Cleanup(func() {
		if err := Init(""); err != nil {
			t.Fatalf("还原配置失败: %v", err)
		}
	})
}

func TestInitWithExplicitFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.ini")
	content := `env_mode = debug
download_lib_path = /tmp/lib

[server]
protocol = https
http_port = 12345
timeout = 3s
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	restore(t)

	if err := Init(path); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if got := EnvMode(); got != "debug" {
		t.Errorf("EnvMode() = %q, want debug", got)
	}
	if got := GetValue("server.protocol"); got != "https" {
		t.Errorf("protocol = %q, want https", got)
	}
	if got := GetInt("server.http_port", 0); got != 12345 {
		t.Errorf("http_port = %d, want 12345", got)
	}
	if got := GetString("download_lib_path", ""); got != "/tmp/lib" {
		t.Errorf("download_lib_path = %q, want /tmp/lib", got)
	}
	if got := GetDuration("server.timeout", time.Second); got != 3*time.Second {
		t.Errorf("timeout = %v, want 3s", got)
	}
}

func TestGettersFallBackToDefault(t *testing.T) {
	restore(t)
	if err := Init(""); err != nil {
		t.Fatal(err)
	}

	if got := GetString("server.not_exist", "默认值"); got != "默认值" {
		t.Errorf("GetString 默认值失效: %q", got)
	}
	if got := GetInt("server.not_exist", 7); got != 7 {
		t.Errorf("GetInt 默认值失效: %d", got)
	}
	if got := GetBool("server.not_exist", true); !got {
		t.Error("GetBool 默认值失效")
	}
	if got := GetDuration("server.not_exist", 2*time.Second); got != 2*time.Second {
		t.Errorf("GetDuration 默认值失效: %v", got)
	}
	// 类型不对时的兜底
	if got := GetInt("server.protocol", 9); got != 9 {
		t.Errorf("非数字配置应当回落到默认值: %d", got)
	}
}

func TestInitMissingExplicitFile(t *testing.T) {
	restore(t)
	if err := Init(filepath.Join(t.TempDir(), "nope.ini")); err == nil {
		t.Fatal("指定了不存在的配置文件，应当返回错误")
	}
}

// TestInitWithEnvPath 验证 GLD_CONFIG 与 -config 等价（配置来源里的第 1 条）。
func TestInitWithEnvPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "from-env.ini")
	if err := os.WriteFile(path, []byte("env_mode = test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	restore(t)
	t.Setenv(PathKey, path)

	if err := Init(""); err != nil {
		t.Fatalf("GLD_CONFIG 指定的配置没被加载: %v", err)
	}
	if got := EnvMode(); got != "test" {
		t.Errorf("GLD_CONFIG 生效后 EnvMode() = %q, want test", got)
	}
}

// TestInitWithMissingEnvPath 验证 GLD_CONFIG 指向不存在的文件时直接报错，
// 而不是悄悄回落到内嵌模板——否则配置名写错了很难发现。
func TestInitWithMissingEnvPath(t *testing.T) {
	restore(t)
	t.Setenv(PathKey, filepath.Join(t.TempDir(), "nope.ini"))

	if err := Init(""); err == nil {
		t.Fatal("GLD_CONFIG 指向不存在的文件，应当返回错误")
	}
}

// TestEmbeddedFallback 验证没有任何外部配置文件时，内嵌模板仍然可用——
// 这是「clone 下来直接 go build 就能跑」的关键。
func TestEmbeddedFallback(t *testing.T) {
	restore(t)

	if err := Init(""); err != nil {
		t.Fatalf("内嵌配置加载失败: %v", err)
	}
	if got := GetString("server.http_port", ""); got != "9801" {
		t.Errorf("内嵌默认端口 = %q, want 9801", got)
	}
	if got := EnvMode(); got != "release" {
		t.Errorf("内嵌默认模式 = %q, want release", got)
	}

	t.Setenv(EnvKey, "debug")
	if err := Init(""); err != nil {
		t.Fatal(err)
	}
	if got := EnvMode(); got != "debug" {
		t.Errorf("GLD_ENV=debug 时 EnvMode() = %q, want debug", got)
	}

	// 非法值应当回落到 release，而不是让 gin.SetMode panic
	t.Setenv(EnvKey, "不存在的环境")
	if err := Init(""); err != nil {
		t.Fatal(err)
	}
	if got := EnvMode(); got != "release" {
		t.Errorf("非法环境时 EnvMode() = %q, want release", got)
	}
}
