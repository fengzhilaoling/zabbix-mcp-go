package zabbix

import (
	"fmt"
)

// Login 登录Zabbix API
func (c *ZabbixClient) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已经设置了token认证，直接验证token有效性
	if c.AuthType == "token" && c.AuthToken != "" {
		// 保存当前的AuthToken，临时清空它以调用不需要认证的API
		savedToken := c.AuthToken
		c.AuthToken = ""

		// 尝试调用apiinfo.version来验证连接是否正常
		_, err := c.call("apiinfo.version", map[string]interface{}{}, "")

		// 恢复AuthToken
		c.AuthToken = savedToken

		if err != nil {
			return fmt.Errorf("token认证失败: %w", err)
		}
		return nil
	}

	// 密码认证
	params := map[string]string{
		"user":     c.User,
		"password": c.Pass,
	}

	response, err := c.call("user.login", params, "")
	if err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	authToken, ok := response.(string)
	if !ok {
		return fmt.Errorf("登录响应格式错误")
	}

	c.AuthToken = authToken
	return nil
}

// Logout 登出Zabbix API
func (c *ZabbixClient) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.AuthToken == "" {
		return nil
	}

	_, err := c.call("user.logout", nil, c.AuthToken)
	c.AuthToken = ""
	return err
}

// SetAuthToken 设置认证token
func (c *ZabbixClient) SetAuthToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AuthToken = token
	c.AuthType = "token"
}

// SetAuthType 设置认证方式
func (c *ZabbixClient) SetAuthType(authType string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AuthType = authType
}

// GetAuthToken 获取当前认证token
func (c *ZabbixClient) GetAuthToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.AuthToken
}

// GetAuthType 获取当前认证方式
func (c *ZabbixClient) GetAuthType() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.AuthType
}
