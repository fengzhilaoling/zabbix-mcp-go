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

// Call 公开API调用方法
func (c *ZabbixClient) Call(method string, params interface{}) (interface{}, error) {
	c.mu.Lock()
	authToken := c.AuthToken
	c.mu.Unlock()

	if authToken == "" {
		if err := c.Login(); err != nil {
			return nil, err
		}
		c.mu.Lock()
		authToken = c.AuthToken
		c.mu.Unlock()
	}

	result, err := c.call(method, params, authToken)
	if err != nil {
		// 如果认证失败，尝试重新登录
		if rpcErr, ok := err.(*RPCError); ok && rpcErr.Code == -32602 {
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
