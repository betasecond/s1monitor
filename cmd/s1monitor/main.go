package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofrs/flock"
	"github.com/s1monitor/pkg/config"
	"github.com/s1monitor/pkg/session"
	"github.com/s1monitor/pkg/tui"
	"github.com/urfave/cli/v2"
)

const (
	// 应用信息
	AppName    = "s1monitor"
	AppVersion = "1.0.0"
	AppUsage   = "S1论坛自动挂机工具"

	// 默认文件路径
	DefaultConfigFile = "config.yaml"
	LockFile          = "s1monitor.lock"
	LogFile           = "s1monitor.log"
)

func main() {
	// 设置应用
	app := &cli.App{
		Name:    AppName,
		Version: AppVersion,
		Usage:   AppUsage,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   DefaultConfigFile,
				Usage:   "指定配置文件路径",
			},
			&cli.BoolFlag{
				Name:    "daemon",
				Aliases: []string{"d"},
				Value:   false,
				Usage:   "在后台守护进程模式运行（无UI）",
			},
		},
		Action: run,
	}

	// 运行应用
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "运行错误: %v\n", err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	// 获取工作目录（程序所在目录）
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}
	workDir := filepath.Dir(exePath)

	// 确定配置文件路径
	configPath := c.String("config")
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(workDir, configPath)
	}

	// 确定锁文件和日志文件路径
	lockPath := filepath.Join(workDir, LockFile)
	logPath := filepath.Join(workDir, LogFile)

	// 设置日志
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("S1Monitor启动 - 版本 %s", AppVersion)

	// 1. 文件锁确保只有一个实例运行
	fileLock := flock.New(lockPath)
	locked, err := fileLock.TryLock()
	if err != nil {
		logger.Printf("获取文件锁失败: %v", err)
		return fmt.Errorf("获取文件锁失败: %v", err)
	}
	if !locked {
		logger.Printf("另一个实例已在运行（无法获取锁 %s）", lockPath)
		return fmt.Errorf("另一个实例已在运行（无法获取锁 %s）", lockPath)
	}
	defer func() {
		fileLock.Unlock()
		logger.Printf("锁文件 %s 已释放", lockPath)
	}()

	logger.Printf("锁文件 %s 获取成功", lockPath)

	// 2. 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// 如果配置文件不存在，尝试创建默认配置
		if os.IsNotExist(err) {
			logger.Printf("配置文件不存在，创建默认配置...")
			if err := config.CreateDefaultConfig(configPath); err != nil {
				logger.Printf("创建默认配置失败: %v", err)
				return fmt.Errorf("创建默认配置失败: %v", err)
			}
			logger.Printf("默认配置已创建在 %s，请编辑后重新启动程序", configPath)
			fmt.Printf("默认配置已创建在 %s，请编辑后重新启动程序\n", configPath)
			return nil
		}
		logger.Printf("加载配置失败: %v", err)
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 3. 创建会话管理器
	sm, err := session.New(cfg, logger)
	if err != nil {
		logger.Printf("创建会话管理器失败: %v", err)
		return fmt.Errorf("创建会话管理器失败: %v", err)
	}

	// 4. 判断模式并运行
	if c.Bool("daemon") {
		// 后台守护进程模式
		logger.Printf("以守护进程模式启动")
		fmt.Println("S1Monitor 已在后台模式启动，查看日志文件获取详情")
		return runDaemon(sm, cfg.Username, logger)
	} else {
		// TUI模式
		logger.Printf("以TUI模式启动")
		monitor := tui.New(sm, cfg.Username, logger)
		monitor.Start()
		return nil
	}
}

// runDaemon 在后台模式运行
func runDaemon(sm *session.SessionManager, username string, logger *log.Logger) error {
	logger.Printf("开始后台监控...")
	logger.Printf("用户名: %s", username)
	logger.Printf("检查间隔: %d 秒", tui.CheckIntervalSeconds)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 用于控制退出的通道
	stopChan := make(chan struct{})

	// 启动监控goroutine
	go func() {
		for {
			select {
			case <-stopChan:
				logger.Printf("监控已停止")
				return
			default:
				// 如果未登录，尝试登录
				if !sm.IsLoggedIn() {
					logger.Printf("尝试登录 (%s)...", username)

					err := sm.Login()
					if err == nil && sm.IsLoggedIn() {
						logger.Printf("登录成功，开始挂机...")
						time.Sleep(time.Duration(tui.CheckIntervalSeconds) * time.Second)
					} else {
						errMsg := ""
						if err != nil {
							errMsg = err.Error()
						} else {
							errMsg = "登录后会话验证失败"
						}
						logger.Printf("登录失败: %s", errMsg)
						logger.Printf("将在 %d 秒后重试...", tui.RetryDelaySeconds)
						time.Sleep(time.Duration(tui.RetryDelaySeconds) * time.Second)
						continue
					}
				} else {
					// 已登录状态，检查会话
					logger.Printf("检查会话有效性...")
					valid, err := sm.CheckSession()
					if err != nil {
						logger.Printf("检查会话时出错: %v", err)
					}

					if valid {
						logger.Printf("会话有效，努力挂机中...")
						time.Sleep(time.Duration(tui.CheckIntervalSeconds) * time.Second)
					} else {
						logger.Printf("会话失效，准备重新登录...")
						time.Sleep(time.Duration(tui.RetryDelaySeconds) * time.Second)
					}
				}
			}
		}
	}()

	// 等待终止信号
	<-sigChan
	logger.Printf("收到终止信号，正在关闭...")
	close(stopChan)

	// 给足够时间让goroutine优雅退出
	time.Sleep(1 * time.Second)
	return nil
}
