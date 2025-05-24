# S1Monitor Go

S1 论坛自动挂机工具（Go 版本）

## 功能特点

- 自动登录 S1 论坛并维持会话
- 支持 TUI（终端用户界面）和后台守护进程模式
- 自动重试登录（当会话失效时）
- 日志记录
- 跨平台支持（Windows、Linux、macOS）
- 自动化发布流程

## 安装

### 从源码编译

确保你已安装 Go 1.16 或更高版本，然后运行：

```bash
git clone https://github.com/betasecond/s1monitor.git
cd s1monitor
go build -o s1monitor ./cmd/s1monitor
```

或者直接使用 `go install`：

```bash
go install github.com/betasecond/s1monitor/cmd/s1monitor@latest
```

### 从发行版下载

你也可以直接从 GitHub Releases 页面下载预编译的二进制文件。

## 配置

首次运行程序时，将自动创建示例配置文件 `config.yaml`。编辑此文件并填入你的用户名和密码：

```yaml
# S1 论坛登录凭据
username: "your_username"
password: "your_password"
```

## 使用方法

### TUI 模式（默认）

```bash
./s1monitor
```

### 后台守护进程模式

```bash
./s1monitor -d
# 或
./s1monitor --daemon
```

### 指定配置文件

```bash
./s1monitor -c /path/to/config.yaml
# 或
./s1monitor --config /path/to/config.yaml
```

## 服务器部署

在 Linux 服务器上，你可以使用 systemd 创建一个服务来管理 S1Monitor：

1. 创建服务文件 `/etc/systemd/system/s1monitor.service`：

```ini
[Unit]
Description=S1 Forum Monitor
After=network.target

[Service]
Type=simple
User=yourusername
ExecStart=/path/to/s1monitor -d
WorkingDirectory=/path/to/s1monitor_dir
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

2. 启用并启动服务：

```bash
sudo systemctl enable s1monitor
sudo systemctl start s1monitor
```

3. 查看日志：

```bash
sudo journalctl -u s1monitor -f
```

## 发布流程

本项目使用 GitHub Actions 实现自动化发布流程，包括以下步骤：

### 创建新版本

1. 为代码库打标签，使用语义化版本号格式：

```bash
git tag -a v1.0.0 -m "发布 v1.0.0 版本"
git push origin v1.0.0
```

2. 推送标签后，GitHub Actions 会自动：
   - 创建 GitHub Release
   - 构建 Windows、Linux 和 macOS 版本的二进制文件
   - 将二进制文件打包并作为资产上传到 Release 页面

### 工作流文件

- `create-release.yml`: 当推送标签时，创建 GitHub Release
- `build-release-binaries.yml`: 当 Release 创建后，构建并上传二进制文件

### 手动构建

如果需要在本地构建多平台二进制文件，可以使用以下命令：

#### Linux/macOS

```bash
make build-all
```

#### Windows

```powershell
.\build.ps1
```

## 许可证

MIT
