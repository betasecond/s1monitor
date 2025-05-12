// Package session 处理S1论坛的登录和会话管理
package session

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/s1monitor/pkg/config"
)

const (
	// 网站URL
	BaseURL   = "https://stage1st.com/2b/"
	LoginURL  = BaseURL + "member.php?mod=logging&action=login&loginsubmit=yes&infloat=yes&lssubmit=yes&inajax=1"
	CheckURL  = BaseURL + "forum.php"
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	Timeout   = 15 * time.Second
)

// SessionManager 管理登录会话
type SessionManager struct {
	config   *config.Config
	client   *http.Client
	loggedIn bool
	logger   *log.Logger
}

// New 创建新的SessionManager
func New(cfg *config.Config, logger *log.Logger) (*SessionManager, error) {
	// 创建cookie jar以处理会话cookie
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建cookie jar失败: %v", err)
	}

	// 配置HTTP客户端
	client := &http.Client{
		Jar:     jar,
		Timeout: Timeout,
	}

	return &SessionManager{
		config:   cfg,
		client:   client,
		loggedIn: false,
		logger:   logger,
	}, nil
}

// Login 执行登录操作
func (sm *SessionManager) Login() error {
	sm.logger.Printf("尝试使用用户名 %s 登录...", sm.config.Username)

	// 准备表单数据
	data := url.Values{
		"fastloginfield": {"username"},
		"username":       {sm.config.Username},
		"password":       {sm.config.Password},
		"quickforward":   {"yes"},
		"handlekey":      {"ls"},
	}

	// 创建请求
	req, err := http.NewRequest("POST", LoginURL, strings.NewReader(data.Encode()))
	if err != nil {
		sm.loggedIn = false
		return fmt.Errorf("创建登录请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := sm.client.Do(req)
	if err != nil {
		sm.loggedIn = false
		return fmt.Errorf("登录请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		sm.loggedIn = false
		return fmt.Errorf("登录请求返回非200状态码: %d", resp.StatusCode)
	}

	// 登录后立即检查会话
	sm.logger.Printf("登录请求已发送，正在验证会话...")
	valid, err := sm.CheckSession()
	if err != nil {
		return fmt.Errorf("验证会话失败: %v", err)
	}

	if valid {
		sm.logger.Printf("用户 %s 登录成功", sm.config.Username)
	} else {
		sm.logger.Printf("登录请求后检查会话失败，可能凭据错误或网站结构变更")
	}

	return nil
}

// CheckSession 检查当前会话是否有效
func (sm *SessionManager) CheckSession() (bool, error) {
	sm.logger.Printf("检查用户 %s 的会话状态...", sm.config.Username)

	// 创建请求
	req, err := http.NewRequest("GET", CheckURL, nil)
	if err != nil {
		sm.loggedIn = false
		return false, fmt.Errorf("创建会话检查请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", UserAgent)

	// 发送请求
	resp, err := sm.client.Do(req)
	if err != nil {
		sm.loggedIn = false
		return false, fmt.Errorf("会话检查请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		sm.loggedIn = false
		return false, fmt.Errorf("会话检查请求返回非200状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sm.loggedIn = false
		return false, fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 检查页面内容是否包含用户名
	content := string(body)
	if strings.Contains(content, sm.config.Username) {
		sm.logger.Printf("会话有效，在页面中找到用户名 %s", sm.config.Username)
		sm.loggedIn = true
		return true, nil
	} else {
		sm.logger.Printf("会话失效或未登录，页面中未找到用户名 %s", sm.config.Username)
		sm.loggedIn = false
		return false, nil
	}
}

// IsLoggedIn 返回当前登录状态
func (sm *SessionManager) IsLoggedIn() bool {
	return sm.loggedIn
}
