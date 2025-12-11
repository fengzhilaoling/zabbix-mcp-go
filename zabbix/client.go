package zabbix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ZabbixClient Zabbix JSON-RPC客户端
type ZabbixClient struct {
	URL        string
	User       string
	Pass       string
	AuthToken  string
	AuthType   string // "password" 或 "token"
	HTTPClient *http.Client
	mu         sync.Mutex
}

// NewZabbixClient 创建新的Zabbix客户端
func NewZabbixClient(url, user, pass string) *ZabbixClient {
	return &ZabbixClient{
		URL:      url,
		User:     user,
		Pass:     pass,
		AuthType: "password", // 默认为密码认证
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// call 调用Zabbix API（内部方法）
func (c *ZabbixClient) call(method string, params interface{}, auth string) (interface{}, error) {
	// 检测Zabbix版本以确定认证方式
	detector := NewVersionDetector(c)
	version, err := detector.DetectVersion()
	if err != nil {
		// 如果版本检测失败，使用传统方式
		return c.callWithAuth(method, params, auth)
	}

	// Zabbix 7.0+ 不再使用auth参数，改用HTTP头部认证
	if version.Major >= 7 {
		return c.callWithHeaderAuth(method, params, auth)
	}

	// 旧版本使用传统的auth参数
	return c.callWithAuth(method, params, auth)
}

// callWithAuth 传统认证方式（Zabbix 6.x及更早版本）
func (c *ZabbixClient) callWithAuth(method string, params interface{}, auth string) (interface{}, error) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
		Auth:    auth,
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构建完整的API URL，如果URL中没有包含api_jsonrpc.php则自动添加
	apiURL := c.URL
	if !strings.Contains(apiURL, "api_jsonrpc.php") {
		// 移除末尾的斜杠（如果有）
		apiURL = strings.TrimRight(apiURL, "/")
		// 添加API路径
		apiURL = apiURL + "/api_jsonrpc.php"
	}

	resp, err := c.HTTPClient.Post(apiURL, "application/json", bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var response JSONRPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// callWithHeaderAuth Zabbix 7.0+ 使用HTTP头部认证
func (c *ZabbixClient) callWithHeaderAuth(method string, params interface{}, auth string) (interface{}, error) {
	// Zabbix 7.0+ 不在请求体中包含auth参数
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
		// Auth字段为空，不包含在JSON中
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构建完整的API URL，如果URL中没有包含api_jsonrpc.php则自动添加
	apiURL := c.URL
	if !strings.Contains(apiURL, "api_jsonrpc.php") {
		// 移除末尾的斜杠（如果有）
		apiURL = strings.TrimRight(apiURL, "/")
		// 添加API路径
		apiURL = apiURL + "/api_jsonrpc.php"
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// Zabbix 7.0+ 使用Authorization头部进行认证
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}

	// 执行请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var response JSONRPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// Call 公开API调用方法
func (c *ZabbixClient) Call(method string, params interface{}) (interface{}, error) {
	c.mu.Lock()
	authToken := c.AuthToken
	c.mu.Unlock()

	if authToken == "" && c.AuthType != "token" {
		if err := c.Login(); err != nil {
			return nil, err
		}
		c.mu.Lock()
		authToken = c.AuthToken
		c.mu.Unlock()
	}

	result, err := c.call(method, params, authToken)
	if err != nil {
		// 如果认证失败，尝试重新登录（仅对密码认证）
		if rpcErr, ok := err.(*RPCError); ok && rpcErr.Code == -32602 && c.AuthType != "token" {
			if err := c.Login(); err != nil {
				return nil, err
			}
			c.mu.Lock()
			authToken = c.AuthToken
			c.mu.Unlock()
			return c.call(method, params, authToken)
		}
		return nil, err
	}

	return result, nil
}

// GetInstanceInfo 获取实例详细信息
func (c *ZabbixClient) GetInstanceInfo() map[string]interface{} {
	info := make(map[string]interface{})
	info["url"] = c.URL
	info["auth_type"] = c.AuthType
	info["status"] = "unknown"

	// 检测版本信息
	detector := NewVersionDetector(c)
	version, err := detector.DetectVersion()
	if err != nil {
		info["version"] = "unknown"
		info["status"] = "version_detection_failed"
	} else {
		info["version"] = version.String()
		info["status"] = "connected"
	}

	return info
}
