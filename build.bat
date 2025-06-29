@echo off
chcp 65001 >nul
:Start_1
echo 请选择需要构建的环境版本
echo 1.debug
echo 2.release
echo 3.test
set /p input_env_mode=请输入数字(1、2、3)，按回车结束：
if %input_env_mode% EQU 1 (
    set select_env_mode=debug
) else if %input_env_mode% EQU 2 (
    set select_env_mode=release
) else if %input_env_mode% EQU 3 (
    set select_env_mode=test
) else (
    echo 输入无效，只能输入1、2、3
    GOTO Start_1
)
echo %select_env_mode%

:Start_2
echo 请选择需要构建的系统平台架构
echo 1、Windows amd64
echo 2、Windows 32
echo 3、Mac amd64
echo 4、Mac arm64
echo 5、Linux amd64
echo 6、Linux 32
echo 7、Linux arm
echo 8、Linux arm64
set /p input_sys=请输入数字(1、2、3、4、5、6、7、8)，按回车结束：
if %input_sys% EQU 1 (
    set "input_sys_goos=windows"
    set "input_sys_goarch=amd64"
) else if %input_sys% EQU 2 (
    set "input_sys_goos=windows"
    set "input_sys_goarch=386"
) else if %input_sys% EQU 3 (
    set "input_sys_goos=darwin"
    set "input_sys_goarch=amd64"
) else if %input_sys% EQU 4 (
    set "input_sys_goos=darwin"
    set "input_sys_goarch=arm64"
) else if %input_sys% EQU 5 (
    set "input_sys_goos=linux"
    set "input_sys_goarch=amd64"
) else if %input_sys% EQU 6 (
    set "input_sys_goos=linux"
    set "input_sys_goarch=386"
) else if %input_sys% EQU 7 (
    set "input_sys_goos=linux"
    set "input_sys_goarch=arm"
) else if %input_sys% EQU 8 (
    set "input_sys_goos=linux"
    set "input_sys_goarch=arm64"
) else (
    echo 输入无效，只能输入1、2、3、4、5、6、7、8
    GOTO Start_2
)

echo 开始构建%input_sys_goos% %input_sys_goarch%程序

cd config
ren "env.ini" "env.ini.temp" && echo 把"env.ini"改为"env.ini.temp"
ren "env.ini.%select_env_mode%" "env.ini" && echo 把"env.ini.%select_env_mode%"改为"env.ini"

echo 编译准备就绪... ...

cd ..

SET CGO_ENABLED=0
SET GOOS=%input_sys_goos%
SET GOARCH=%input_sys_goarch%

for %%i in (%cd%) do set folder_name=%%~nxi
set app_path=build\output\%select_env_mode%\%GOOS%_%GOARCH%
set app_name=%folder_name%-%GOOS%-%GOARCH%
set is_build_app=0

IF "%input_sys_goos%" == "windows" GOTO Windows
IF "%input_sys_goos%" == "darwin" GOTO LinuxMac
IF "%input_sys_goos%" == "linux" GOTO LinuxMac
GOTO End

:Windows
echo 请耐心等待......
go build -o %app_path%\%app_name%.exe && (
    echo @echo off > %app_path%\start.bat
    echo .\%app_name%.exe >> %app_path%\start.bat
    echo pause >> %app_path%\start.bat
) && echo 编译结束 && set is_build_app=1
GOTO End

:LinuxMac
echo 请耐心等待......
set "sh_start=chmod +x ./%app_name% ^&^& ./%app_name%"
go build -o %app_path%\%app_name% && (
    echo #!/bin/sh > %app_path%\start.sh
    echo %sh_start% >> %app_path%\start.sh
) && echo 编译结束 && set is_build_app=1
GOTO End

:End
cd config
ren "env.ini" "env.ini.%select_env_mode%" && echo 把"env.ini"改回"env.ini.%select_env_mode%"
ren "env.ini.temp" "env.ini" && echo 把"env.ini.temp"改回"env.ini"

echo 构建结束
IF %is_build_app% EQU 1 echo 文件存放在%app_path%目录里

pause