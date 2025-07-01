### Step 1 创建一个资源存放目录
  `mkdir /root/download_lib`
### Step 2 拉取镜像
  `docker pull tutudev99/golocaldownload:25.06.29.10`
#### Or 或者从阿里云拉取
  `docker pull registry.cn-shenzhen.aliyuncs.com/tutudev99/golocaldownload:25.06.29.10`  

  `docker tag registry.cn-shenzhen.aliyuncs.com/tutudev99/golocaldownload:25.06.29.10 tutudev99/golocaldownload:25.06.29.10`  

  `docker rmi registry.cn-shenzhen.aliyuncs.com/tutudev99/golocaldownload:25.06.29.10`
### Step 3 运行容器
  `docker run
    -p 9801:9801
    --name golocaldownload 
    -v /root/download_lib:/root/download_lib
    --restart always
    -d tutudev99/golocaldownload:25.06.29.10
  `
### 说明：
#### 访问地址就是 `http://本地服务器IP:9801/`，提供下载的文件就放在本地服务器`/root/download_lib`目录下