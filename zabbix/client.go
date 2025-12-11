package zabbix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// JSONRPCRequest JSON-RPC请求结构
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
	Auth    string      `json:"auth,omitempty"`
}

// JSONRPCResponse JSON-RPC响应结构
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *RPCError   `json:"error"`
	ID      int         `json:"id"`
}

// RPCError RPC错误结构
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("Zabbix API Error %d: %s (%s)", e.Code, e.Message, e.Data)
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

// Login 登录Zabbix API
func (c *ZabbixClient) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已经设置了token认证，直接验证token有效性
	if c.AuthType == "token" && c.AuthToken != "" {
		// 尝试调用一个简单的API来验证token是否有效
		_, err := c.call("apiinfo.version", nil, c.AuthToken)
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

// call 调用Zabbix API
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

	zabbix_url := c.URL + "/api_jsonrpc.php"
	resp, err := c.HTTPClient.Post(zabbix_url, "application/json", bytes.NewBuffer(requestData))
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

// GetHosts 获取主机列表
func (c *ZabbixClient) GetHosts(groupID, hostName string) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output": []string{"hostid", "host", "name", "status", "available"},
	}

	if groupID != "" {
		params["groupids"] = groupID
	}

	if hostName != "" {
		params["filter"] = map[string]string{
			"host": hostName,
		}
	}

	result, err := c.Call("host.get", params)
	if err != nil {
		return nil, err
	}

	hosts, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var hostList []map[string]interface{}
	for _, h := range hosts {
		if host, ok := h.(map[string]interface{}); ok {
			hostList = append(hostList, host)
		}
	}

	return hostList, nil
}

// GetHostByName 根据主机名获取主机信息
func (c *ZabbixClient) GetHostByName(hostName string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"output": "extend",
		"filter": map[string]string{
			"host": hostName,
		},
		"selectGroups":     "extend",
		"selectInterfaces": "extend",
	}

	result, err := c.Call("host.get", params)
	if err != nil {
		return nil, err
	}

	hosts, ok := result.([]interface{})
	if !ok || len(hosts) == 0 {
		return nil, fmt.Errorf("主机不存在")
	}

	if host, ok := hosts[0].(map[string]interface{}); ok {
		return host, nil
	}

	return nil, fmt.Errorf("响应格式错误")
}

// CreateHost 创建主机
func (c *ZabbixClient) CreateHost(hostName, groupID, interfaceIP string) (string, error) {
	params := map[string]interface{}{
		"host": hostName,
		"interfaces": []map[string]interface{}{
			{
				"type":  1,
				"main":  1,
				"useip": 1,
				"ip":    interfaceIP,
				"dns":   "",
				"port":  "10050",
			},
		},
		"groups": []map[string]interface{}{
			{
				"groupid": groupID,
			},
		},
	}

	result, err := c.Call("host.create", params)
	if err != nil {
		return "", err
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误")
	}

	hostIDs, ok := response["hostids"].([]interface{})
	if !ok || len(hostIDs) == 0 {
		return "", fmt.Errorf("创建主机失败")
	}

	return hostIDs[0].(string), nil
}

// DeleteHost 删除主机
func (c *ZabbixClient) DeleteHost(hostID string) error {
	params := []string{hostID}
	_, err := c.Call("host.delete", params)
	return err
}

// GetItems 获取主机监控项
func (c *ZabbixClient) GetItems(hostID string) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":             "extend",
		"hostids":            hostID,
		"selectApplications": "extend",
		"selectTriggers":     "extend",
	}

	result, err := c.Call("item.get", params)
	if err != nil {
		return nil, err
	}

	items, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var itemList []map[string]interface{}
	for _, i := range items {
		if item, ok := i.(map[string]interface{}); ok {
			itemList = append(itemList, item)
		}
	}

	return itemList, nil
}

// GetItemData 获取监控项数据
func (c *ZabbixClient) GetItemData(itemID string, history, limit int) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":    "extend",
		"history":   history,
		"itemids":   itemID,
		"sortfield": "clock",
		"sortorder": "DESC",
		"limit":     limit,
	}

	method := "history.get"
	if history == 0 {
		method = "history.get"
	}

	result, err := c.Call(method, params)
	if err != nil {
		return nil, err
	}

	data, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var dataList []map[string]interface{}
	for _, d := range data {
		if item, ok := d.(map[string]interface{}); ok {
			dataList = append(dataList, item)
		}
	}

	return dataList, nil
}

// CreateItem 创建监控项
func (c *ZabbixClient) CreateItem(hostID, name, key, type_, valueType, delay string) (string, error) {
	if type_ == "" {
		type_ = "0" // Zabbix agent
	}
	if valueType == "" {
		valueType = "3" // Numeric unsigned
	}
	if delay == "" {
		delay = "60s"
	}

	params := map[string]interface{}{
		"name":        name,
		"key_":        key,
		"hostid":      hostID,
		"type":        type_,
		"value_type":  valueType,
		"delay":       delay,
		"interfaceid": "$1", // 使用默认接口
	}

	result, err := c.Call("item.create", params)
	if err != nil {
		return "", err
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误")
	}

	itemIDs, ok := response["itemids"].([]interface{})
	if !ok || len(itemIDs) == 0 {
		return "", fmt.Errorf("创建监控项失败")
	}

	return itemIDs[0].(string), nil
}

// GetTriggers 获取触发器
func (c *ZabbixClient) GetTriggers(hostID string, active bool) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":      "extend",
		"selectHosts": "extend",
		"selectItems": "extend",
	}

	if hostID != "" {
		params["hostids"] = hostID
	}

	if active {
		params["filter"] = map[string]interface{}{
			"status": 0,
		}
	}

	result, err := c.Call("trigger.get", params)
	if err != nil {
		return nil, err
	}

	triggers, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var triggerList []map[string]interface{}
	for _, t := range triggers {
		if trigger, ok := t.(map[string]interface{}); ok {
			triggerList = append(triggerList, trigger)
		}
	}

	return triggerList, nil
}

// GetTriggerEvents 获取触发器事件
func (c *ZabbixClient) GetTriggerEvents(triggerID string, limit int) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":              "extend",
		"select_acknowledges": "extend",
		"objectids":           triggerID,
		"source":              0,
		"object":              0,
		"sortfield":           "clock",
		"sortorder":           "DESC",
		"limit":               limit,
	}

	result, err := c.Call("event.get", params)
	if err != nil {
		return nil, err
	}

	events, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var eventList []map[string]interface{}
	for _, e := range events {
		if event, ok := e.(map[string]interface{}); ok {
			eventList = append(eventList, event)
		}
	}

	return eventList, nil
}

// AcknowledgeEvent 确认事件
func (c *ZabbixClient) AcknowledgeEvent(eventID, message string) error {
	params := map[string]interface{}{
		"eventids": eventID,
		"message":  message,
	}

	_, err := c.Call("event.acknowledge", params)
	return err
}

// GetTemplates 获取模板列表
func (c *ZabbixClient) GetTemplates() ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":       []string{"templateid", "host", "name", "description"},
		"selectGroups": "extend",
	}

	result, err := c.Call("template.get", params)
	if err != nil {
		return nil, err
	}

	templates, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var templateList []map[string]interface{}
	for _, t := range templates {
		if template, ok := t.(map[string]interface{}); ok {
			templateList = append(templateList, template)
		}
	}

	return templateList, nil
}

// LinkTemplate 关联模板到主机
func (c *ZabbixClient) LinkTemplate(hostID, templateID string) error {
	params := map[string]interface{}{
		"hosts": []map[string]string{
			{"hostid": hostID},
		},
		"templates": []map[string]string{
			{"templateid": templateID},
		},
	}

	_, err := c.Call("template.massadd", params)
	return err
}

// UnlinkTemplate 从主机移除模板
func (c *ZabbixClient) UnlinkTemplate(hostID, templateID string) error {
	params := map[string]interface{}{
		"hostids":           hostID,
		"templateids_clear": templateID,
	}

	_, err := c.Call("host.massremove", params)
	return err
}
