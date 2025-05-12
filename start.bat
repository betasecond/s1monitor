@echo off
REM S1Monitor 启动脚本 - Windows版本

echo 正在启动 S1 论坛监控工具...
echo.

REM 检测程序是否存在
if not exist s1monitor.exe (
    echo 错误: 未找到 s1monitor.exe 程序。
    echo 请确保您已经编译了程序或者位于正确的目录。
    echo.
    pause
    exit /b 1
)

REM 启动方式菜单
echo 请选择启动方式:
echo 1. 交互界面模式
echo 2. 后台守护进程模式
echo.

set /p choice="请输入数字 (1/2): "

if "%choice%"=="1" (
    echo 正在以交互界面模式启动...
    start s1monitor.exe
) else if "%choice%"=="2" (
    echo 正在以后台守护进程模式启动...
    start s1monitor.exe -d
) else (
    echo 无效选择，默认以交互界面模式启动...
    start s1monitor.exe
)

echo.
echo 启动完成。
echo 如果需要查看日志，请检查 s1monitor.log 文件。
echo.
pause
