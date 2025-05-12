// Package tui 处理终端用户界面
package tui

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/s1monitor/pkg/session"
)

// 常量定义
const (
	LogBufferSize        = 1000 // 日志缓冲区大小
	CheckIntervalSeconds = 60   // 检查间隔，60秒
	RetryDelaySeconds    = 10   // 登录失败或检查失败后的重试延迟
)

// StatusType 定义当前状态类型
type StatusType int

// 状态类型常量
const (
	StatusUnknown StatusType = iota
	StatusInitializing
	StatusLoggingIn
	StatusLoginSuccess
	StatusLoginFailed
	StatusSessionValid
	StatusSessionInvalid
)

// 状态文本映射
var statusText = map[StatusType]string{
	StatusUnknown:        "未知",
	StatusInitializing:   "初始化...",
	StatusLoggingIn:      "登录中...",
	StatusLoginSuccess:   "登录成功",
	StatusLoginFailed:    "登录失败",
	StatusSessionValid:   "会话有效",
	StatusSessionInvalid: "会话失效",
}

// Monitor 是S1监控的TUI应用
type Monitor struct {
	app            *tview.Application
	pages          *tview.Pages
	logView        *tview.TextView
	statusBar      *tview.TextView
	sessionManager *session.SessionManager
	logs           []string
	currentStatus  StatusType
	username       string
	stopChan       chan struct{}
	logger         *log.Logger
}

// New 创建新的TUI应用
func New(sm *session.SessionManager, username string, logger *log.Logger) *Monitor {
	// 创建应用实例
	app := tview.NewApplication()
	pages := tview.NewPages()

	// 创建日志视图
	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetScrollable(true).
		ScrollToEnd()
	logView.SetBorder(true).SetTitle(" 日志 ")

	// 创建状态栏
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// 创建监视器实例
	m := &Monitor{
		app:            app,
		pages:          pages,
		logView:        logView,
		statusBar:      statusBar,
		sessionManager: sm,
		logs:           make([]string, 0, LogBufferSize),
		currentStatus:  StatusInitializing,
		username:       username,
		stopChan:       make(chan struct{}),
		logger:         logger,
	}

	// 创建主布局
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(logView, 0, 1, false).
		AddItem(statusBar, 1, 0, false)

	// 将布局添加到页面
	pages.AddPage("main", flex, true, true)

	// 设置应用的根元素
	app.SetRoot(pages, true).EnableMouse(true)

	// 设置键盘事件处理
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Ctrl+C 退出
		if event.Key() == tcell.KeyCtrlC {
			m.Stop()
			return nil
		}
		return event
	})

	return m
}

// Log 添加日志消息到日志视图
func (m *Monitor) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("15:04:05")
	logMsg := fmt.Sprintf("[%s] %s", timestamp, msg)

	// 添加到日志数组
	m.logs = append(m.logs, logMsg)
	if len(m.logs) > LogBufferSize {
		m.logs = m.logs[1:]
	}

	// 更新日志视图
	m.app.QueueUpdateDraw(func() {
		fmt.Fprintln(m.logView, logMsg)
	})

	// 记录到文件日志
	m.logger.Println(msg)
}

// UpdateStatus 更新当前状态
func (m *Monitor) UpdateStatus(status StatusType) {
	m.currentStatus = status

	// 更新状态栏
	statusStr := statusText[status]
	loginStatus := "未登录"
	if m.sessionManager.IsLoggedIn() {
		loginStatus = "[green]已登录[-]"
	} else if status == StatusLoggingIn {
		loginStatus = "[yellow]登录中[-]"
	} else if status == StatusLoginFailed || status == StatusSessionInvalid {
		loginStatus = "[red]登录失败[-]"
	}

	statusBarText := fmt.Sprintf("状态: %s | 登录: %s | 用户: %s | %s",
		statusStr, loginStatus, m.username, time.Now().Format("15:04:05"))

	m.app.QueueUpdateDraw(func() {
		m.statusBar.Clear()
		fmt.Fprint(m.statusBar, statusBarText)
	})

	// 同时记录到日志
	m.Log("状态更新: %s", statusStr)
}

// MonitorSession 后台监控会话
func (m *Monitor) MonitorSession() {
	m.Log("[green]开始会话监控[white]")
	m.Log("用户名: %s", m.username)
	m.Log("检查间隔: %d 秒", CheckIntervalSeconds)

	for {
		select {
		case <-m.stopChan:
			m.Log("[red]会话监控已停止[white]")
			return
		default:
			// 如果未登录，尝试登录
			if !m.sessionManager.IsLoggedIn() {
				m.UpdateStatus(StatusLoggingIn)
				m.Log("尝试登录 (%s)...", m.username)

				err := m.sessionManager.Login()
				if err == nil && m.sessionManager.IsLoggedIn() {
					m.UpdateStatus(StatusLoginSuccess)
					m.Log("登录成功，开始挂机...")
					time.Sleep(time.Duration(CheckIntervalSeconds) * time.Second)
				} else {
					m.UpdateStatus(StatusLoginFailed)
					errMsg := ""
					if err != nil {
						errMsg = err.Error()
					} else {
						errMsg = "登录后会话验证失败"
					}
					m.Log("登录失败: %s", errMsg)
					m.Log("将在 %d 秒后重试...", RetryDelaySeconds)
					time.Sleep(time.Duration(RetryDelaySeconds) * time.Second)
					continue
				}
			} else {
				// 已登录状态，检查会话
				m.Log("检查会话有效性...")
				valid, err := m.sessionManager.CheckSession()
				if err != nil {
					m.Log("检查会话时出错: %v", err)
				}

				if valid {
					m.UpdateStatus(StatusSessionValid)
					m.Log("会话有效，努力挂机中...")
					time.Sleep(time.Duration(CheckIntervalSeconds) * time.Second)
				} else {
					m.UpdateStatus(StatusSessionInvalid)
					m.Log("会话失效，准备重新登录...")
					time.Sleep(time.Duration(RetryDelaySeconds) * time.Second)
				}
			}
		}
	}
}

// Start 启动应用
func (m *Monitor) Start() {
	// 启动会话监控（在后台goroutine中）
	go m.MonitorSession()

	// 运行应用
	if err := m.app.Run(); err != nil {
		m.logger.Printf("应用运行出错: %v", err)
	}
}

// Stop 停止应用
func (m *Monitor) Stop() {
	m.Log("[red]收到退出信号，正在关闭...[white]")
	close(m.stopChan)
	m.app.Stop()
}
