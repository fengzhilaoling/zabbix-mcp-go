package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"zabbix-mcp-go/zabbix"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"
)

// Config 多实例配置
type Config struct {
	Instances []ZabbixInstance `yaml:"instances"`
}

// ZabbixInstance Zabbix实例配置
type ZabbixInstance struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	User     string `yaml:"username,omitempty"`
	Pass     string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
	AuthType string `yaml:"auth_type,omitempty"` // "password" 或 "token"
	Default  bool   `yaml:"default,omitempty"`
}

var (
	config Config
	pool   *zabbix.ZabbixPool
)

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化Zabbix连接池
	pool = zabbix.NewZabbixPool()
	for _, instance := range config.Instances {
		client := zabbix.NewZabbixClient(instance.URL, instance.User, instance.Pass)

		// 设置认证方式
		if instance.AuthType == "token" && instance.Token != "" {
			client.SetAuthToken(instance.Token)
		}

		if err := pool.AddInstance(instance.Name, client); err != nil {
			log.Printf("添加实例 %s 失败: %v", instance.Name, err)
			continue
		}
		if instance.Default {
			pool.SetDefault(instance.Name)
		}
	}

	// 创建MCP服务器
	s := server.NewMCPServer(
		"zabbix-mcp-server",
		"1.0.0",
	)

	// 注册工具
	registerTools(s)

	// 启动服务器
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func loadConfig() error {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

func registerTools(s *server.MCPServer) {
	// 主机相关工具
	s.AddTool(
		mcp.NewTool("get_hosts",
			mcp.WithDescription("获取Zabbix主机列表"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("group_id", mcp.Description("主机组ID")),
			mcp.WithString("host_name", mcp.Description("主机名筛选")),
		),
		getHostsHandler,
	)
	s.AddTool(
		mcp.NewTool("get_host_by_name",
			mcp.WithDescription("根据主机名获取主机信息"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_name", mcp.Required(), mcp.Description("主机名")),
		),
		getHostByNameHandler,
	)
	s.AddTool(
		mcp.NewTool("create_host",
			mcp.WithDescription("创建Zabbix主机"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_name", mcp.Required(), mcp.Description("主机名")),
			mcp.WithString("group_id", mcp.Required(), mcp.Description("主机组ID")),
			mcp.WithString("interface_ip", mcp.Required(), mcp.Description("接口IP地址")),
		),
		createHostHandler,
	)
	s.AddTool(
		mcp.NewTool("delete_host",
			mcp.WithDescription("删除Zabbix主机"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
		),
		deleteHostHandler,
	)

	// 监控项相关工具
	s.AddTool(
		mcp.NewTool("get_items",
			mcp.WithDescription("获取主机监控项"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
		),
		getItemsHandler,
	)
	s.AddTool(
		mcp.NewTool("get_item_data",
			mcp.WithDescription("获取监控项数据"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("item_id", mcp.Required(), mcp.Description("监控项ID")),
			mcp.WithNumber("history", mcp.DefaultNumber(0), mcp.Description("历史数据类型")),
			mcp.WithNumber("limit", mcp.DefaultNumber(100), mcp.Description("数据条数限制")),
		),
		getItemDataHandler,
	)
	s.AddTool(
		mcp.NewTool("create_item",
			mcp.WithDescription("创建监控项"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("监控项名称")),
			mcp.WithString("key", mcp.Required(), mcp.Description("监控项键值")),
			mcp.WithString("type", mcp.Description("监控项类型")),
			mcp.WithString("value_type", mcp.Description("值类型")),
			mcp.WithString("delay", mcp.Description("更新间隔")),
		),
		createItemHandler,
	)

	// 触发器相关工具
	s.AddTool(
		mcp.NewTool("get_triggers",
			mcp.WithDescription("获取触发器列表"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Description("主机ID")),
			mcp.WithBoolean("active", mcp.Description("只显示启用的触发器")),
		),
		getTriggersHandler,
	)
	s.AddTool(
		mcp.NewTool("get_trigger_events",
			mcp.WithDescription("获取触发器事件"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("trigger_id", mcp.Required(), mcp.Description("触发器ID")),
			mcp.WithNumber("limit", mcp.DefaultNumber(100), mcp.Description("事件数量限制")),
		),
		getTriggerEventsHandler,
	)
	s.AddTool(
		mcp.NewTool("acknowledge_event",
			mcp.WithDescription("确认事件"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("event_id", mcp.Required(), mcp.Description("事件ID")),
			mcp.WithString("message", mcp.Description("确认消息")),
		),
		acknowledgeEventHandler,
	)

	// 模板相关工具
	s.AddTool(
		mcp.NewTool("get_templates",
			mcp.WithDescription("获取模板列表"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
		),
		getTemplatesHandler,
	)
	s.AddTool(
		mcp.NewTool("link_template",
			mcp.WithDescription("关联模板到主机"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
			mcp.WithString("template_id", mcp.Required(), mcp.Description("模板ID")),
		),
		linkTemplateHandler,
	)
	s.AddTool(
		mcp.NewTool("unlink_template",
			mcp.WithDescription("从主机移除模板"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
			mcp.WithString("template_id", mcp.Required(), mcp.Description("模板ID")),
		),
		unlinkTemplateHandler,
	)

	// 多实例管理
	s.AddTool(
		mcp.NewTool("list_instances",
			mcp.WithDescription("列出所有Zabbix实例"),
		),
		listInstancesHandler,
	)
	s.AddTool(
		mcp.NewTool("switch_instance",
			mcp.WithDescription("切换当前使用的实例"),
			mcp.WithString("instance", mcp.Required(), mcp.Description("实例名称")),
		),
		switchInstanceHandler,
	)
}

// 工具处理器示例
func getHostsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	groupID := ""
	hostName := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["group_id"].(string); ok {
		groupID = v
	}
	if v, ok := args["host_name"].(string); ok {
		hostName = v
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	hosts, err := client.GetHosts(groupID, hostName)
	if err != nil {
		return nil, fmt.Errorf("获取主机列表失败: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", hosts),
			},
		},
	}, nil
}

func getHostByNameHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	hostName := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_name"].(string); ok {
		hostName = v
	}

	if hostName == "" {
		return nil, fmt.Errorf("主机名不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	host, err := client.GetHostByName(hostName)
	if err != nil {
		return nil, fmt.Errorf("获取主机信息失败: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", host),
			},
		},
	}, nil
}
func createHostHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	hostName := ""
	groupID := ""
	interfaceIP := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_name"].(string); ok {
		hostName = v
	}
	if v, ok := args["group_id"].(string); ok {
		groupID = v
	}
	if v, ok := args["interface_ip"].(string); ok {
		interfaceIP = v
	}

	if hostName == "" || groupID == "" || interfaceIP == "" {
		return nil, fmt.Errorf("主机名、组ID和接口IP不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	hostID, err := client.CreateHost(hostName, groupID, interfaceIP)
	if err != nil {
		return nil, fmt.Errorf("创建主机失败: %v", err)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"hostid":  hostID,
		"message": fmt.Sprintf("主机 %s 创建成功", hostName),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}
func deleteHostHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	hostID := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}

	if hostID == "" {
		return nil, fmt.Errorf("主机ID不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	if err := client.DeleteHost(hostID); err != nil {
		return nil, fmt.Errorf("删除主机失败: %v", err)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("主机 %s 删除成功", hostID),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

func getItemsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	hostID := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}

	if hostID == "" {
		return nil, fmt.Errorf("主机ID不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	items, err := client.GetItems(hostID)
	if err != nil {
		return nil, fmt.Errorf("获取监控项失败: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", items),
			},
		},
	}, nil
}

func getItemDataHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	itemID := ""
	history := 0
	limit := 100

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["item_id"].(string); ok {
		itemID = v
	}
	if v, ok := args["history"].(float64); ok {
		history = int(v)
	}
	if v, ok := args["limit"].(float64); ok {
		limit = int(v)
	}

	if itemID == "" {
		return nil, fmt.Errorf("监控项ID不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	data, err := client.GetItemData(itemID, history, limit)
	if err != nil {
		return nil, fmt.Errorf("获取监控项数据失败: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", data),
			},
		},
	}, nil
}

func createItemHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	hostID := ""
	name := ""
	key := ""
	type_ := ""
	valueType := ""
	delay := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["name"].(string); ok {
		name = v
	}
	if v, ok := args["key"].(string); ok {
		key = v
	}
	if v, ok := args["type"].(string); ok {
		type_ = v
	}
	if v, ok := args["value_type"].(string); ok {
		valueType = v
	}
	if v, ok := args["delay"].(string); ok {
		delay = v
	}

	if hostID == "" || name == "" || key == "" {
		return nil, fmt.Errorf("主机ID、监控项名称和键值不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	itemID, err := client.CreateItem(hostID, name, key, type_, valueType, delay)
	if err != nil {
		return nil, fmt.Errorf("创建监控项失败: %v", err)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"itemid":  itemID,
		"message": fmt.Sprintf("监控项 %s 创建成功", name),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

func getTriggersHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	hostID := ""
	active := true

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["active"].(bool); ok {
		active = v
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	triggers, err := client.GetTriggers(hostID, active)
	if err != nil {
		return nil, fmt.Errorf("获取触发器失败: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", triggers),
			},
		},
	}, nil
}

func getTriggerEventsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	triggerID := ""
	limit := 100

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["trigger_id"].(string); ok {
		triggerID = v
	}
	if v, ok := args["limit"].(float64); ok {
		limit = int(v)
	}

	if triggerID == "" {
		return nil, fmt.Errorf("触发器ID不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	events, err := client.GetTriggerEvents(triggerID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取触发器事件失败: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", events),
			},
		},
	}, nil
}

func acknowledgeEventHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	eventID := ""
	message := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["event_id"].(string); ok {
		eventID = v
	}
	if v, ok := args["message"].(string); ok {
		message = v
	}

	if eventID == "" {
		return nil, fmt.Errorf("事件ID不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	if err := client.AcknowledgeEvent(eventID, message); err != nil {
		return nil, fmt.Errorf("确认事件失败: %v", err)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("事件 %s 确认成功", eventID),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

func getTemplatesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	templates, err := client.GetTemplates()
	if err != nil {
		return nil, fmt.Errorf("获取模板失败: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", templates),
			},
		},
	}, nil
}

func linkTemplateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	hostID := ""
	templateID := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["template_id"].(string); ok {
		templateID = v
	}

	if hostID == "" || templateID == "" {
		return nil, fmt.Errorf("主机ID和模板ID不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	if err := client.LinkTemplate(hostID, templateID); err != nil {
		return nil, fmt.Errorf("关联模板失败: %v", err)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("模板 %s 关联到主机 %s 成功", templateID, hostID),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

func unlinkTemplateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""
	hostID := ""
	templateID := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["template_id"].(string); ok {
		templateID = v
	}

	if hostID == "" || templateID == "" {
		return nil, fmt.Errorf("主机ID和模板ID不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	if err := client.UnlinkTemplate(hostID, templateID); err != nil {
		return nil, fmt.Errorf("移除模板失败: %v", err)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("模板 %s 从主机 %s 移除成功", templateID, hostID),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

func listInstancesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instances := pool.ListInstances()
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", instances),
			},
		},
	}, nil
}

func switchInstanceHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	instanceName := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}

	if instanceName == "" {
		return nil, fmt.Errorf("实例名称不能为空")
	}

	if err := pool.SetDefault(instanceName); err != nil {
		return nil, fmt.Errorf("切换实例失败: %v", err)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("已切换到实例: %s", instanceName),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

// 辅助函数
func getStringParam(params map[string]interface{}, key string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return ""
}

func getIntParam(params map[string]interface{}, key string, defaultVal int) int {
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

func getBoolParam(params map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := params[key].(bool); ok {
		return val
	}
	return defaultVal
}
