@echo off
rem Launcher for golocaldownload.exe.
rem
rem Keep this file next to the binary (the release archives already ship it):
rem double-click it, or run "start.bat", to launch in the foreground. Every
rem argument is passed straight through to the program:
rem
rem     start.bat                      run with the embedded defaults
rem     start.bat -config .\env.ini    use an external config file
rem
rem Ctrl+C stops it - in-flight downloads are allowed to finish first. After the
rem program exits the window stays open so the exit code can be read; set
rem GLD_NO_PAUSE=1 to skip that when calling this script from another script.
rem
rem NOTE: keep this file ASCII-only. cmd.exe reads batch files with the OEM code
rem page (GBK on Chinese Windows), so UTF-8 text here would show up as garbage.
rem It is also CRLF-only - see .gitattributes.

setlocal
cd /d "%~dp0"

if not exist "golocaldownload.exe" (
    echo start.bat: golocaldownload.exe not found in %~dp0
    echo            Keep this script in the same directory as the executable.
    if not "%GLD_NO_PAUSE%"=="1" pause
    exit /b 1
)

echo Starting golocaldownload.exe ... ^(Ctrl+C to stop^)
echo.
golocaldownload.exe %*
set "code=%ERRORLEVEL%"
echo.
echo golocaldownload.exe exited with code %code%.
if not "%GLD_NO_PAUSE%"=="1" pause
exit /b %code%
