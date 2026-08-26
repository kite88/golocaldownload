开发语言：Go 

开发框架：Gin

# 文档

| 仓库    | 地址                                       |
| ------- | ------------------------------------------ |
| GitHub  | https://github.com/kite88/golocaldownload  |
| GitCode | https://gitcode.com/kite88/golocaldownload |
| Gitee   | https://gitee.com/kite88/golocaldownload   |

## 方式一：本地源码部署、二次开发

### Step 1 代码拉下来 (以GitCode为例)

```
git clone https://gitcode.com/kite88/golocaldownload.git
```

### Step 2 把config文件下的env.ini.local改为env.ini

说明：

env.ini文件是配置文件，例如，端口配置、存放下载资源目录配置

## 方式二：本地主机 运行可执行文件

### Step 1 下载发行版，找到对应自己系统平台架构的压缩包下载

- [github下载地址](https://github.com/kite88/golocaldownload/releases/tag/25.07.07.00)
- [gitcode下载地址](https://gitcode.com/kite88/golocaldownload/releases/25.07.07.00)
- [gitee下载地址](https://gitee.com/kite88/golocaldownload/releases/tag/25.07.07.00)

### Step 2 解压后会有可执行程序跟启动脚本

说明：

访问地址就是 `http://本地服务器IP:9801/`，

提供下载的资源就放在当前程序执行的路径 `download_lib` 目录下

## 方式三：docker部署（得有docker环境）

### Step 1 代码拉下来 (以GitCode为例)

```
git clone https://gitcode.com/kite88/golocaldownload.git
```

### Step 2 进入到项目根目录直接命令行依次输入以下两条指令

#### 构建

```
docker build -t golocaldownload:25.07.07.00 .
```

#### 运行

```
docker run -p 9801:9801 -v /home/download_lib:/root/download_lib --restart always --name golocaldownload-app -d golocaldownload:25.07.07.00
```

参数说明：

- -p 9801:9801，将本地主机的 9801 端口映射到容器内的 9801 端口
- -v /home/download\_lib:/root/download\_lib，将主机的 /home/download\_lib 目录挂载到容器内的 /root/download\_lib 目录（如果你是windows系统且安装了Linux子系统及Docker Desktop，也可以用windows目录，如: -v D:\\download\_lib:/root/download\_lib）
- --restart always，设置容器的重启策略为always，即使容器停止也会自动重启
- --name golocaldownload-app，将容器命名为 golocaldownload-app
- -d 在后台运行容器并返回容器 ID

说明：

访问地址就是 `http://本地服务器IP:9801/`，

提供下载的资源就放在 `/home/download_lib` 目录下

## 方式四：docker-compose工具部署（得有docker、docker-compose环境）

### Step 1 代码拉下来（GitCode为例）

```
git clone https://gitcode.com/kite88/golocaldownload.git
```

### Step 2 进入到项目根目录直接命令行输入

```
docker-compose up -d
```

说明：

根据docker-compose.yml的配置，端口为 9801，存放下载资源目录为 /home/download\_lib

访问地址就是 `http://本地服务器IP:9801/`，

提供下载的资源就放在 `/home/download_lib` 目录下

## 方式五：docker制品直接部署

### Step 1 拉取镜像

```
docker pull tutudev99/golocaldownload:25.07.07.00
```

温馨提示：如果由于网络环境拉不下来可以采用 [方式六：阿里云制品仓库拉取部署方式](#%E6%96%B9%E5%BC%8F%E5%85%AD%EF%BC%9A%E9%98%BF%E9%87%8C%E4%BA%91%E5%88%B6%E5%93%81%E4%BB%93%E5%BA%93%E6%8B%89%E5%8F%96%E9%83%A8%E7%BD%B2%E6%96%B9%E5%BC%8F)

### Step 2 运行容器

```
docker run -p 9801:9801 --name golocaldownload -v /home/download_lib:/root/download_lib --restart always -d tutudev99/golocaldownload:25.07.07.00
```

参数说明：

- -p 9801:9801，将本地主机的 9801 端口映射到容器内的 9801 端口
- --name golocaldownload-app，将容器命名为 golocaldownload-app
- -v /home/download\_lib:/root/download\_lib，将主机的 /home/download\_lib 目录挂载到容器内的 /root/download\_lib 目录（如果你是windows系统且安装了Linux子系统及Docker Desktop，也可以用windows目录，如: -v D:\\download\_lib:/root/download\_lib）
- --restart always，设置容器的重启策略为always，即使容器停止也会自动重启
- -d 在后台运行容器并返回容器 ID

说明：

访问地址就是 `http://本地服务器IP:9801/`，

提供下载的资源就放在 `/home/download_lib` 目录下

## 方式六：阿里云制品仓库拉取部署方式

### Step 1 从阿里云仓库拉取镜像

```
docker pull registry.cn-shenzhen.aliyuncs.com/tutudev99/golocaldownload:25.07.07.00
```

### Step 2 运行容器

```
docker run -p 9801:9801 --name golocaldownload -v /home/download_lib:/root/download_lib --restart always -d registry.cn-shenzhen.aliyuncs.com/tutudev99/golocaldownload:25.07.07.00
```

参数说明：

- -p 9801:9801，将本地主机的 9801 端口映射到容器内的 9801 端口
- --name golocaldownload-app，将容器命名为 golocaldownload-app
- -v /home/download\_lib:/root/download\_lib，将主机的 /home/download\_lib 目录挂载到容器内的 /root/download\_lib 目录（如果你是windows系统且安装了Linux子系统及Docker Desktop，也可以用windows目录，如: -v D:\\download\_lib:/root/download\_lib）
- --restart always，设置容器的重启策略为always，即使容器停止也会自动重启
- -d 在后台运行容器并返回容器 ID

说明：

访问地址就是 `http://本地服务器IP:9801/`，

提供下载的资源就放在 `/home/download_lib` 目录下

