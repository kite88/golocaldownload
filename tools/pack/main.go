// Command pack 把可执行文件（可再带上若干附加文件，比如启动脚本）打成 .tar.gz 或 .zip，
// 并显式写入归档内的文件权限。
//
// 为什么不用系统工具：
//   - Windows 自带的 tar.exe 是精简版 bsdtar：它不读文件权限（Windows 上没有 Unix
//     执行位这个概念，它一律伪造 0666），也不支持 --mode，打出来的包在 Linux 上
//     解压后是 644、根本执行不了；
//   - zip(1) 在 macOS / Linux 上并非必然存在，不同实现写出的字节也不一致。
//
// Go 标准库可以显式设置权限、并把时间戳固定下来，因此无论在哪个平台、用
// build.ps1 还是 build.sh，打出的包都完全一致（可复现、校验和相同）。
// 格式由 -out 的扩展名决定：.zip 出 zip，其余出 tar.gz。
//
// 用法（由构建脚本调用，一般不需要手动执行）：
//
//	pack -in dist/golocaldownload -out dist/golocaldownload-linux-amd64.tar.gz -name golocaldownload -mode 0755 -add start.sh=/repo/start.sh
//	pack -in dist/golocaldownload.exe -out dist/golocaldownload-windows-amd64.zip -name golocaldownload.exe -add start.bat=/repo/start.bat
//
// -add 可重复，写法是「归档内的文件名=源文件路径」；归档里的顺序就是 -in、-add 的
// 书写顺序，附加文件一律 0755（start.sh 必须可执行；start.bat 在 Windows 上不看权限位，
// 多出来的执行位无害），这样同一份源码在任何平台打出的包逐字节一致。
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// addMode 是 -add 附加文件的权限：启动脚本要能在 Unix 上直接执行，
// 所以统一带上执行位（Windows 侧不看这个位）。
const addMode int64 = 0o755

// entry 是归档里的一个文件：name 是归档内的名字，path 是源文件路径。
type entry struct {
	name string
	path string
	mode int64
}

// addList 收集可重复的 -add 参数（flag 包只提供了单值 flag，需要自己实现 Value）。
type addList []string

func (a *addList) String() string { return strings.Join(*a, ", ") }

func (a *addList) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	in := flag.String("in", "", "输入文件（必填）")
	out := flag.String("out", "", "输出路径，.zip 或 .tar.gz（必填）")
	name := flag.String("name", "", "归档内的文件名，默认取输入文件名")
	mode := flag.String("mode", "0755", "归档内的文件权限，八进制，默认 0755")
	var adds addList
	flag.Var(&adds, "add", "附加文件，写成「归档内文件名=源文件路径」，可重复；权限固定 0755")
	flag.Parse()

	entries, err := buildEntries(*in, *name, *mode, adds)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pack:", err)
		os.Exit(1)
	}
	if err := pack(*out, entries); err != nil {
		fmt.Fprintln(os.Stderr, "pack:", err)
		os.Exit(1)
	}
}

// buildEntries 把命令行参数整理成归档条目列表，并把参数错误挡在写文件之前。
func buildEntries(in, name, modeStr string, adds []string) ([]entry, error) {
	if in == "" {
		return nil, fmt.Errorf("-in 为必填")
	}
	mode, err := strconv.ParseInt(modeStr, 8, 32)
	if err != nil {
		return nil, fmt.Errorf("解析 -mode 失败: %w", err)
	}
	if name == "" {
		name = filepath.Base(in)
	}

	entries := []entry{{name: name, path: in, mode: mode}}
	seen := map[string]bool{name: true}

	for _, spec := range adds {
		addName, addPath, ok := strings.Cut(spec, "=")
		if !ok || addName == "" || addPath == "" {
			return nil, fmt.Errorf("-add 需要写成「归档内文件名=源文件路径」: %q", spec)
		}
		// 同名条目会让归档里出现两个同名文件：解压时后者覆盖前者，很难排查，直接拒绝。
		if seen[addName] {
			return nil, fmt.Errorf("-add 里的文件名重复: %q", addName)
		}
		seen[addName] = true
		entries = append(entries, entry{name: addName, path: addPath, mode: addMode})
	}
	return entries, nil
}

func pack(out string, entries []entry) error {
	if out == "" {
		return fmt.Errorf("-out 为必填")
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	if strings.EqualFold(filepath.Ext(out), ".zip") {
		err = writeZip(f, entries)
	} else {
		err = writeTarGz(f, entries)
	}
	if err != nil {
		f.Close()
		os.Remove(out) // 失败时别留下半个包
		return err
	}
	return f.Close()
}

// writeTarGz 写 .tar.gz。时间戳固定为 Unix 0 点：源码不变时打出的包完全一致，
// 因此在 tar -tvf 里看到 1970-01-01 是有意为之。
func writeTarGz(w io.Writer, entries []entry) error {
	gz := gzip.NewWriter(w) // ModTime 同样留零值
	tw := tar.NewWriter(gz)

	for _, e := range entries {
		src, err := os.Open(e.path)
		if err != nil {
			return fmt.Errorf("%s: %w", e.path, err)
		}
		fi, err := src.Stat()
		if err != nil {
			src.Close()
			return fmt.Errorf("%s: %w", e.path, err)
		}

		// Uid/Gid/Uname/Gname 不设置：避免把打包机的用户身份写进归档
		hdr := &tar.Header{
			Name:     e.name,
			Typeflag: tar.TypeReg,
			Mode:     e.mode,
			Size:     fi.Size(),
			ModTime:  time.Unix(0, 0),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			src.Close()
			return err
		}
		if _, err := io.Copy(tw, src); err != nil {
			src.Close()
			return err
		}
		if err := src.Close(); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// writeZip 写 .zip。SetMode 会同时设置「version made by = Unix」和外置属性，
// 这样解压工具才会认权限位。
func writeZip(w io.Writer, entries []entry) error {
	zw := zip.NewWriter(w)

	for _, e := range entries {
		src, err := os.Open(e.path)
		if err != nil {
			return fmt.Errorf("%s: %w", e.path, err)
		}

		hdr := &zip.FileHeader{
			Name:   e.name,
			Method: zip.Deflate,
			// 时间戳固定为 zip 的纪元（1980-01-01）：DOS 时间字段留 0 时，
			// 不少工具会显示成 1979-11-30，写死更清楚，同样保证可复现。
			Modified: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		hdr.SetMode(fs.FileMode(e.mode))

		fw, err := zw.CreateHeader(hdr)
		if err != nil {
			src.Close()
			return err
		}
		if _, err := io.Copy(fw, src); err != nil {
			src.Close()
			return err
		}
		if err := src.Close(); err != nil {
			return err
		}
	}

	return zw.Close()
}
