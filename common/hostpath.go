package common

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// mountInfoFile 是内核维护的挂载表。容器里能读到它，但直接跑二进制时（Windows / macOS）
// 这个路径不存在，此时无从探测，返回空串即可。
const mountInfoFile = "/proc/self/mountinfo"

// HostPath 尽力推断 root 在宿主机上的对应路径，推断不出来时返回空串（调用方回落到 root 本身）。
//
// 为什么需要它：容器里看到的永远是容器内路径（如 /root/download_lib），而用户关心的其实是
// 宿主机上那个被映射的目录（如 D:\download_lib）——「往哪儿放文件」才是页面上要回答的问题。
//
// 做法是读 /proc/self/mountinfo：每次 bind mount 都会留一行，第 4 个字段是「该文件系统内的
// 路径」，对 bind mount 来说正是宿主机上的目录。取挂载点是 root 前缀且最长的那条记录，
// 拼上相对部分即可。各平台实测差异：
//
//   - Linux 原生、macOS 的 Docker Desktop：第 4 个字段本身就是宿主机路径；
//   - Windows 的 Docker Desktop 走 9p/drvfs：第 4 个字段**没有盘符**（实测形如
//     /Users/ASUS/AppData/Local/Temp/x），盘符藏在超级选项的 aname=drvfs;path=C:\ 里，需要补回。
//
// 命名卷（named volume）的挂载根是 "/"，拼出来的结果没有意义，按「推断不出来」处理。
func HostPath(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(mountInfoFile)
	if err != nil {
		return ""
	}
	return hostPathIn(string(data), abs)
}

// hostPathIn 从 mountinfo 内容里推断 abs 对应的宿主机路径，是 HostPath 的可测部分。
func hostPathIn(mountInfo, abs string) string {
	best, bestLen := "", -1
	for _, line := range strings.Split(mountInfo, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		mountPoint := unescapeMountPath(fields[4])
		if mountPoint == "/" {
			// 容器的根文件系统（overlay）那一条挂载点是 "/"，它会匹配上任何路径，
			// 但它的「挂载根」指的是容器自己的根，拼不出宿主机路径。
			// 库没落在任何 bind mount 里时，就老实回落到容器内路径。
			continue
		}
		if !insidePath(mountPoint, abs) || len(mountPoint) <= bestLen {
			continue
		}
		p, ok := hostPathOf(fields, mountPoint, abs)
		if ok {
			best, bestLen = p, len(mountPoint)
		}
	}
	return best
}

// insidePath 判断绝对路径 dir 是否落在 mountPoint 之内（含两者相等）。
func insidePath(mountPoint, dir string) bool {
	if mountPoint == dir {
		return true
	}
	return strings.HasPrefix(dir, strings.TrimSuffix(mountPoint, "/")+"/")
}

// hostPathOf 由一条 mountinfo 记录拼出宿主机路径：宿主机侧挂载根 + 相对挂载点的部分。
// 第二个返回值为 false 表示这条记录给不出有意义的宿主机路径（命名卷等）。
func hostPathOf(fields []string, mountPoint, abs string) (string, bool) {
	hostRoot := unescapeMountPath(fields[3])
	rel := strings.TrimPrefix(strings.TrimPrefix(abs, mountPoint), "/")

	// Windows 的 Docker Desktop：第 4 个字段没有盘符，从超级选项里取回来
	if drive := windowsDrive(fields); drive != "" {
		full := hostRoot
		if rel != "" {
			full += "/" + rel
		}
		return drive + strings.ReplaceAll(full, "/", `\`), true
	}

	// 用 path 而不是 filepath：mountinfo 里的路径永远是 POSIX 风格，
	// filepath 在 Windows 上会把分隔符换成反斜杠，拼出来的路径就不对了
	full := path.Join(hostRoot, rel)
	if full == "/" || full == "." {
		// 挂载根本身就是 "/"：命名卷之类，报个 "/" 出去只会误导人
		return "", false
	}
	return full, true
}

// windowsDrive 从 mountinfo 记录的超级选项里取出 Windows 盘符（如 path=C:\ 或 path=C:\134）。
// 只有 Docker Desktop 的 9p/drvfs 才有这个选项，其它文件系统返回空串。
func windowsDrive(fields []string) string {
	for i, field := range fields {
		if field != "-" || i+3 >= len(fields) {
			continue
		}
		for _, opt := range strings.Split(fields[i+3], ";") {
			v, ok := strings.CutPrefix(opt, "path=")
			if ok && len(v) >= 2 && v[1] == ':' {
				return v[:2]
			}
		}
	}
	return ""
}

// unescapeMountPath 还原 mountinfo 里的八进制转义：空格 \040、制表符 \011、
// 换行 \012、反斜杠 \134（路径里带空格时全靠这个）。
func unescapeMountPath(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) && isOctal(s[i+1]) && isOctal(s[i+2]) && isOctal(s[i+3]) {
			b.WriteByte((s[i+1]-'0')<<6 | (s[i+2]-'0')<<3 | (s[i+3] - '0'))
			i += 3
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isOctal(c byte) bool { return c >= '0' && c <= '7' }
