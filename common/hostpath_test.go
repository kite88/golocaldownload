package common

import (
	"os"
	"testing"
)

// 下面两条是实测拿到的 mountinfo 片段（原样保留转义），用来固定解析行为。
//
// Linux 原生 bind mount：第 4 个字段就是宿主机路径。
const mountLinux = `25 30 0:24 / / rw,relatime - overlay overlay rw,lowerdir=/var/lib/docker/overlay2/x
477 468 0:68 /home/download_lib /root/download_lib rw,noatime - ext4 /dev/sda1 rw,errors=continue
`

// Windows 的 Docker Desktop：走 9p/drvfs，第 4 个字段没有盘符，盘符在超级选项里。
const mountWindowsDesktop = `25 30 0:24 / / rw,relatime - overlay overlay rw
477 468 0:68 /Users/ASUS/AppData/Local/Temp/gld-mnt /root/download_lib rw,noatime - 9p C:\134 rw,aname=drvfs;path=C:\;uid=0;gid=0;metadata;symlinkroot=/mnt/host/,cache=5,access=client,msize=65536,trans=fd,rfd=5,wfd=5
`

func TestHostPathIn(t *testing.T) {
	tests := []struct {
		name      string
		mountInfo string
		abs       string
		want      string
	}{
		{"Linux 原生 bind mount", mountLinux, "/root/download_lib", "/home/download_lib"},
		{"Windows Docker Desktop", mountWindowsDesktop, "/root/download_lib", `C:\Users\ASUS\AppData\Local\Temp\gld-mnt`},
		{
			name:      "Windows Docker Desktop（盘符写成八进制转义）",
			mountInfo: "477 468 0:68 /Users/me/lib /root/download_lib rw,noatime - 9p C:\\134 rw,aname=drvfs;path=C:\\134;uid=0\n",
			abs:       "/root/download_lib",
			want:      `C:\Users\me\lib`,
		},
		{
			name:      "库是挂载点的子目录",
			mountInfo: "477 468 0:68 /home/data /root rw,noatime - ext4 /dev/sda1 rw\n",
			abs:       "/root/download_lib",
			want:      "/home/data/download_lib",
		},
		{
			name:      "路径里有空格（八进制转义 \\040）",
			mountInfo: "477 468 0:68 /home/my\\040lib /root/download_lib rw - ext4 /dev/sda1 rw\n",
			abs:       "/root/download_lib",
			want:      "/home/my lib",
		},
		{
			name:      "命名卷：挂载根是 /，给不出有意义的宿主机路径",
			mountInfo: "478 468 0:69 / /root/download_lib rw,noatime - ext4 /dev/sdb rw\n",
			abs:       "/root/download_lib",
			want:      "",
		},
		{
			name:      "没有任何匹配的挂载点",
			mountInfo: mountLinux,
			abs:       "/data/download_lib",
			want:      "",
		},
		{
			name:      "空内容",
			mountInfo: "",
			abs:       "/root/download_lib",
			want:      "",
		},
		{
			name: "多条匹配时取最长挂载点",
			mountInfo: "10 1 0:20 / / rw - ext4 /dev/sda1 rw\n" +
				"477 468 0:68 /host/data /root rw - ext4 /dev/sda1 rw\n" +
				"478 468 0:69 /host/data/lib /root/download_lib rw - ext4 /dev/sda1 rw\n",
			abs:  "/root/download_lib",
			want: "/host/data/lib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hostPathIn(tt.mountInfo, tt.abs); got != tt.want {
				t.Errorf("hostPathIn() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestHostPathOutsideContainer 验证非容器环境（读不到 mountinfo）不会报错、只会返回空串，
// 调用方据此回落到真实的下载库路径。
func TestHostPathOutsideContainer(t *testing.T) {
	if _, err := os.Stat(mountInfoFile); err == nil {
		t.Skip("当前环境能读到 /proc/self/mountinfo，跳过")
	}
	if got := HostPath("/tmp/whatever"); got != "" {
		t.Errorf("读不到 mountinfo 时应返回空串，得到 %q", got)
	}
}

// TestUnescapeMountPath 单独覆盖转义还原。
func TestUnescapeMountPath(t *testing.T) {
	tests := []struct{ in, want string }{
		{"/home/a b", "/home/a b"},
		{`/home/my\040lib`, "/home/my lib"},
		{`/a\011b`, "/a\tb"},
		{`/a\134b`, `/a\b`},
		{`/a\04`, `/a\04`}, // 位数不足，原样保留
	}
	for _, tt := range tests {
		if got := unescapeMountPath(tt.in); got != tt.want {
			t.Errorf("unescapeMountPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
