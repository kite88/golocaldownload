# golocaldownload · 本地文件下载服务

[![Release](https://img.shields.io/github/v/release/kite88/golocaldownload)](https://github.com/kite88/golocaldownload/releases/latest)
[![License](https://img.shields.io/github/license/kite88/golocaldownload)](LICENSE)

用 Go + Gin 写的本地文件下载服务：把服务器上的某个目录当作「下载库」，浏览器里就能浏览目录、
按文件名全局检索、点击下载。页面模板与静态资源全部内嵌进二进制，发布出去就是一个可执行文件，
配置也有内嵌默认值，**解压即用**。

| 仓库    | 地址                                       |
| ------- | ------------------------------------------ |
| GitHub  | https://github.com/kite88/golocaldownload  |
| GitCode | https://gitcode.com/kite88/golocaldownload |
| Gitee   | https://gitee.com/kite88/golocaldownload   |

## 功能

- **目录浏览**：面包屑逐级跳转，目录排在文件前面，展示大小与修改时间。
- **全局检索**：按文件名做忽略大小写的子串匹配，可一键定位到文件所在目录。
- **文件下载**：中文、空格等非 ASCII 文件名按 RFC 5987 编码，浏览器不会存成乱码。
- **越界防护**：列表、检索、下载都只允许访问下载库之内的路径，`../` 之类的目录穿越会被直接拒绝。
- **单文件交付**：页面模板、静态资源、配置模板全部内嵌，运行时不依赖任何外部文件。
- **优雅退出**：`Ctrl+C` 或 `SIGTERM` 会等待在途下载收尾后再退出，不会掐断大文件传输。

## 目录结构

```text
.
├── main.go                 程序入口：装配配置、路由、优雅退出
├── config/                 配置加载（内嵌 env.ini.<环境> 模板 + 外部 env.ini 覆盖）
├── common/                 通用工具：目录准备、路径安全解析、体积格式化
├── handle/                 三个接口的实现：列表 / 检索 / 下载
├── router/                 路由装配：页面、静态资源、API 分组
├── test/                   进程内路由集成测试
├── web/
│   ├── view/index.html     页面模板（内嵌）
│   └── static/             Bootstrap / jQuery / 图标 / 前端脚本（内嵌）
├── tools/pack/             归档打包器：让 zip 与 tar.gz 可复现（构建脚本调用）
├── build.ps1 / build.sh    交叉编译全部平台
└── Dockerfile              两阶段构建的最小镜像
```

## 快速开始

需要 Go 1.23 及以上。

```bash
# 开发模式直接跑
go run .

# 或编译后运行
go build -o golocaldownload . && ./golocaldownload
```

启动后会打印版本、配置来源、下载库目录与可访问地址，访问其中任意一个地址即可：

```text
2026-09-28 10:00:00 golocaldownload dev（debug）
配置来源: config/env.ini
下载库目录: E:\download_lib
本机访问: http://127.0.0.1:9801
局域网访问: http://192.168.1.10:9801
```

把要对外提供的文件丢进「下载库目录」，刷新页面就能看到。

### 命令行参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-config` | 空 | 指定外部配置文件（ini）路径；留空时按下面的优先级自动查找 |
| `-version` | `false` | 打印版本号后退出 |

### 配置项

配置文件是 INI 格式（`config/env.ini.<环境>` 是仓库里的模板）：

| 配置项 | 默认值 | 说明 |
| --- | --- | --- |
| `env_mode` | `release` | 运行模式：`debug` / `release` / `test`，取值非法时按 `release` 处理 |
| `download_lib_path` | `download_lib` | 下载库目录；支持相对与绝对路径，多级目录会自动创建；留空表示用程序当前工作目录 |
| `server.protocol` | `http` | 仅用于启动日志里拼出访问地址，不改变实际监听方式 |
| `server.http_port` | `9801` | 监听端口 |

配置来源按优先级从高到低：

1. `-config` 参数 / `GLD_CONFIG` 环境变量指定的文件（指定了就必须存在）；
2. 工作目录下的 `env.ini`；
3. 工作目录下的 `config/env.ini`；
4. 二进制内嵌的 `env.ini.<GLD_ENV>`（`GLD_ENV` 默认 `release`）；
5. 二进制内嵌的 `env.ini.release`。

> 相比旧版，`download_lib_path` 现在允许写 `./a/b` 这类多级路径（会逐级创建），
> 并且**不再需要**为了编译而把 `env.ini.local` 改名成 `env.ini`：默认走内嵌模板，
> 源码 clone 下来直接 `go build` 即可。

## 交叉编译

```powershell
.\build.ps1                      # Windows（PowerShell 5.1 / 7）
.\build.ps1 -OutDir D:\tmp\dist  # 换输出目录
```

```bash
./build.sh                       # macOS / Linux
./build.sh -o /tmp/gld-dist      # 换输出目录
```

目标矩阵写在两个脚本的顶部，增删一行即可。默认覆盖：

| 系统 | 架构 |
| --- | --- |
| Windows | amd64 / arm64 / 386 |
| Linux | amd64 / arm64 / 386 / arm(v7) / ppc64le / s390x / riscv64 / loong64 |
| macOS | amd64 / arm64 |
| FreeBSD | amd64 / arm64 |

两个脚本跑同一份目标矩阵、调用同一个打包工具，**产物逐字节一致**（校验和相同），默认输出到
`dist/`（已被 `.gitignore` 忽略）。按业界惯例，Windows 出 zip，其余平台出 `.tar.gz`
（zip 不保存 Unix 执行位，tar.gz 才能让 Linux / macOS 用户解压后直接运行）：

```text
dist/
├── golocaldownload-windows-amd64.zip    →  golocaldownload.exe
├── golocaldownload-linux-arm64.tar.gz   →  golocaldownload  （带 0755 执行位）
├── ……                                    （共 15 个平台）
├── golocaldownload(.exe)                ← 本机平台的未压缩版，本机自用、不发布
└── checksums.txt                        ← 15 个归档的 SHA256（LF 换行，可 sha256sum -c）
```

归档里的可执行文件带版本号（取自 `git describe`，非 git 环境为 `dev`），`golocaldownload -version` 可查。

> 归档由 `tools/pack` 生成，而不是调用系统 `tar` / `zip`：Windows 自带的 `tar.exe`（精简版
> bsdtar）不读 Unix 权限位、也不支持 `--mode`，打出来的包在 Linux 上解压是 644、跑不起来；
> `zip(1)` 在 macOS / Linux 上也不是必然存在，而且各实现写出的字节不一致。`pack` 用 Go 标准库
> 显式写入 0755，并把时间戳固定下来，所以同源码无论在哪个平台、用哪个脚本构建，产物和校验和都可复现。

### 发版

推 `v*` 或日期式标签（如 `25.07.07.01`）即可，workflow（`.github/workflows/release.yml`）
会自动跑静态检查、单元测试、交叉编译全部平台并创建 GitHub Release：

```bash
git tag 25.07.07.01 && git push origin 25.07.07.01
```

流程会用 `sha256sum -c checksums.txt` 自检产物，任一步失败都不会发版；标签名含 `-`（如
`25.07.07.01-rc1`）会自动标记为 prerelease，不会顶掉 Latest。

## 部署

### 方式一：本地源码部署、二次开发

```bash
git clone https://gitcode.com/kite88/golocaldownload.git
cd golocaldownload
go run .
```

需要自定义配置时，把 `config/` 下的某个 `env.ini.<环境>` 复制成 `config/env.ini`
（或者直接 `go run . -config config/env.ini.local`），不必改动模板文件本身。

### 方式二：本地主机运行可执行文件

从 [Releases](https://github.com/kite88/golocaldownload/releases/latest) 下载对应平台的压缩包，
解压后直接运行；如果需要换端口或下载库目录，在同目录放一份 `env.ini` 即可。

### 方式三：Docker 部署

```bash
docker build -t golocaldownload:latest .
docker run -p 9801:9801 -v /home/download_lib:/root/download_lib --restart always --name golocaldownload-app -d golocaldownload:latest
```

参数说明：

- `-p 9801:9801`，把宿主机的 9801 端口映射到容器的 9801 端口；
- `-v /home/download_lib:/root/download_lib`，把宿主机的下载库目录挂载到容器内（Windows 装了
  Linux 子系统与 Docker Desktop，也可以写成 `-v D:\download_lib:/root/download_lib`）；
- `--restart always`，容器退出后自动重启；
- `-d`，后台运行。

### 方式四：docker-compose 部署

```bash
docker compose up -d      # 旧版 docker-compose 命令为 docker-compose up -d
```

端口为 9801，下载库目录为 `/home/download_lib`，可在 `docker-compose.yml` 中修改。

### 方式五：Docker 制品直接部署

```bash
docker pull tutudev99/golocaldownload:latest
docker run -p 9801:9801 --name golocaldownload -v /home/download_lib:/root/download_lib --restart always -d tutudev99/golocaldownload:latest
```

### 方式六：阿里云制品仓库拉取部署

```bash
docker pull registry.cn-shenzhen.aliyuncs.com/tutudev99/golocaldownload:latest
docker run -p 9801:9801 --name golocaldownload -v /home/download_lib:/root/download_lib --restart always -d registry.cn-shenzhen.aliyuncs.com/tutudev99/golocaldownload:latest
```

> 两个仓库都由发版流程自动推送两份标签：`latest` 与对应版本号（如 `26.09.28.00`），镜像同时覆盖
> linux/amd64 与 linux/arm64。需要固定版本、避免 `latest` 随发版漂移时，把命令里的 `:latest`
> 换成具体版本号即可。

## 接口

| 方法与路径 | 参数 | 说明 |
| --- | --- | --- |
| `GET /` | - | 页面 |
| `GET /api/list` | `path`（相对下载库的路径，如 `/sub`，留空为根目录） | 列出目录内容 |
| `POST /api/search` | 表单 `keyword` | 按文件名检索，最多返回 500 条 |
| `GET /api/download` | `data`（文件路径的 base64url 编码） | 下载文件 |

出错时返回对应的状态码与 `{"error": "..."}`：路径越界 `400` / `403`，目标不存在 `404`，
目录不可读 `403`，其他内部错误 `500`。

`GET /api/list` 的响应：

```json
{
  "root_dir": "/data/download_lib",
  "absolute_dir": "/data/download_lib/sub",
  "relative_dirs": [{"": ""}, {"sub": "/sub"}],
  "list": [
    {
      "name": "a.zip",
      "is_dir": false,
      "size": 1.5,
      "size_unit": "MB",
      "mod_time": "2026/09/28 10:00",
      "parent_path": "/sub",
      "path": "/sub/a.zip",
      "pathname_key": "L2RhdGEvZG93bmxvYWRfbGliL3N1Yi9hLnppcA=="
    }
  ]
}
```

## 实现要点

- **所有请求路径都经过越界校验**：旧实现把 `rootDir` 与 `path` 直接字符串拼接
  （`rootDir + path`），传 `?path=/../../etc/passwd` 就能列出下载库之外的目录，
  下载接口更是可以读取服务端任意文件。现在统一走 `common.SafeJoin`：先拒绝含 NUL 的路径，
  再统一分隔符、清掉首部斜杠，最后用 `filepath.Rel` 确认结果没有逃出根目录；
  下载接口还会额外校验「绝对路径也必须落在根目录之内」。
- **路由常开 `Recovery`**：旧实现在 release 模式用 `gin.New()`（既没有 Logger 也没有
  Recovery），而 handler 里用 `log.Panicln` 处理错误，一次「目录不存在」就能让进程直接退出。
  现在 handler 只返回状态码与错误信息，引擎在任何模式下都挂着 `Recovery`。
- **配置内嵌四份环境模板**：旧实现只 `//go:embed env.ini`，而 `env.ini` 被 `.gitignore`
  排除、仓库里并不存在，clone 下来 `go build` 会报 `pattern env.ini: no matching files found`，
  必须先手工改名。现在内嵌 `env.ini.debug / .local / .release / .test`，用 `GLD_ENV` 选择，
  外部 `env.ini` 依然可以覆盖，编译开箱即用。
- **配置访问带默认值且不会 panic**：`env_mode` 取值非法时回落到 `release`，避免 `gin.SetMode`
  直接 panic；`GetInt` / `GetBool` / `GetDuration` 在配置缺失或写错时返回默认值。
- **下载文件名按 RFC 5987 编码**：用标准库 `mime.FormatMediaType` 生成 `Content-Disposition`，
  中文文件名不再乱码。
- **检索用 `filepath.WalkDir` 并限制条数**：`WalkDir` 不做多余的 `lstat`、不跟随符号链接；
  单个条目读不到时跳过，不影响整体检索；返回上限 500 条，避免一次检索把响应体撑爆。
- **目录排在文件前面**：服务端排好序再返回，前端不用再排一遍。
- **静态资源长缓存**：页面、图标随二进制发布，`Cache-Control: public, max-age=86400`；
  另外把根路径的 `/favicon.ico` 也指向内嵌图标。
- **优雅退出**：`signal.NotifyContext` + `http.Server.Shutdown`，只限制读请求头超时，
  下载大文件时不会被写超时掐断。
- **根目录显式注入**：不再用 `GLD_download_lib_path` 环境变量在包之间隐式传递，
  `handle.New(root)` 让接口天然可测——`test/` 里用 `httptest` 在进程内跑真实路由，
  覆盖了列表、检索、下载、越界拒绝与 404 等场景。

## 已知局限

1. **没有鉴权**：任何能访问到端口的人都能浏览与下载下载库里的全部内容，默认只适合本机或内网使用。
2. **没有 HTTPS**：`server.protocol` 只影响启动日志里打印的地址，真正的 TLS 需要由前置
   Nginx / Caddy 之类的反向代理终结。
3. **检索是全量遍历**：每次请求都会 `WalkDir` 一遍下载库，没建索引；下载库特别大时会偏慢。
4. **只支持一个下载库根目录**，不支持多目录映射，也没有上传、删除、重命名等写操作。
5. **没有断点续传与限速**：下载走 `gin.Context.File`，由标准库 `http.ServeContent` 处理
   Range 请求，但没有做并发数限制与限速，大文件高并发时容易把带宽打满。
6. **前端仍是 jQuery + 字符串拼 HTML**（已做转义），没有做前后端分离与构建链路。

## 开发

```bash
go vet ./...    # 静态检查
go test ./...   # 单元测试 + 进程内路由集成测试
```

改动接口或路径处理逻辑后，请务必跑一遍 `test/router_test.go`，其中的越界用例是安全回归测试。

## 许可证

[MIT](LICENSE)
