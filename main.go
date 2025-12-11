package main

import (
	"context"
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
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	User    string `yaml:"user"`
	Pass    string `yaml:"pass"`
	Default bool   `yaml:"default,omitempty"`
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
			mcp.WithBoolean("active", mcp.DefaultBoolean(true), mcp.Description("只显示启用的触发器")),
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
func getHostsHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	groupID := getStringParam(params, "group_id")
	hostName := getStringParam(params, "host_name")

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	hosts, err := client.GetHosts(groupID, hostName)
	if err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("获取主机列表失败: %v", err)), nil
	}

	return mcp.NewToolResult(hosts), nil
}

func getHostByNameHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	hostName := getStringParam(params, "host_name")
	if hostName == "" {
		return mcp.NewErrorResult("主机名不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	host, err := client.GetHostByName(hostName)
	if err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("获取主机信息失败: %v", err)), nil
	}

	return mcp.NewToolResult(host), nil
}

func createHostHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	hostName := getStringParam(params, "host_name")
	groupID := getStringParam(params, "group_id")
	interfaceIP := getStringParam(params, "interface_ip")

	if hostName == "" || groupID == "" || interfaceIP == "" {
		return mcp.NewErrorResult("主机名、组ID和接口IP不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	hostID, err := client.CreateHost(hostName, groupID, interfaceIP)
	if err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("创建主机失败: %v", err)), nil
	}

	return mcp.NewToolResult(map[string]interface{}{
		"hostid":  hostID,
		"message": fmt.Sprintf("主机 %s 创建成功", hostName),
	}), nil
}

func deleteHostHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	hostID := getStringParam(params, "host_id")
	if hostID == "" {
		return mcp.NewErrorResult("主机ID不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	if err := client.DeleteHost(hostID); err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("删除主机失败: %v", err)), nil
	}

	return mcp.NewToolResult(map[string]interface{}{
		"message": fmt.Sprintf("主机 %s 删除成功", hostID),
	}), nil
}

func getItemsHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	hostID := getStringParam(params, "host_id")
	if hostID == "" {
		return mcp.NewErrorResult("主机ID不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	items, err := client.GetItems(hostID)
	if err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("获取监控项失败: %v", err)), nil
	}

	return mcp.NewToolResult(items), nil
}

func getItemDataHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	itemID := getStringParam(params, "item_id")
	history := getIntParam(params, "history", 0)
	limit := getIntParam(params, "limit", 100)

	if itemID == "" {
		return mcp.NewErrorResult("监控项ID不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	data, err := client.GetItemData(itemID, history, limit)
	if err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("获取监控项数据失败: %v", err)), nil
	}

	return mcp.NewToolResult(data), nil
}

func createItemHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	hostID := getStringParam(params, "host_id")
	name := getStringParam(params, "name")
	key := getStringParam(params, "key")
	type_ := getStringParam(params, "type")
	valueType := getStringParam(params, "value_type")
	delay := getStringParam(params, "delay")

	if hostID == "" || name == "" || key == "" {
		return mcp.NewErrorResult("主机ID、监控项名称和键值不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	itemID, err := client.CreateItem(hostID, name, key, type_, valueType, delay)
	if err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("创建监控项失败: %v", err)), nil
	}

	return mcp.NewToolResult(map[string]interface{}{
		"itemid":  itemID,
		"message": fmt.Sprintf("监控项 %s 创建成功", name),
	}), nil
}

func getTriggersHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	hostID := getStringParam(params, "host_id")
	active := getBoolParam(params, "active", true)

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	triggers, err := client.GetTriggers(hostID, active)
	if err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("获取触发器失败: %v", err)), nil
	}

	return mcp.NewToolResult(triggers), nil
}

func getTriggerEventsHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	triggerID := getStringParam(params, "trigger_id")
	limit := getIntParam(params, "limit", 100)

	if triggerID == "" {
		return mcp.NewErrorResult("触发器ID不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	events, err := client.GetTriggerEvents(triggerID, limit)
	if err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("获取触发器事件失败: %v", err)), nil
	}

	return mcp.NewToolResult(events), nil
}

func acknowledgeEventHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	eventID := getStringParam(params, "event_id")
	message := getStringParam(params, "message")

	if eventID == "" {
		return mcp.NewErrorResult("事件ID不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	if err := client.AcknowledgeEvent(eventID, message); err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("确认事件失败: %v", err)), nil
	}

	return mcp.NewToolResult(map[string]interface{}{
		"message": fmt.Sprintf("事件 %s 确认成功", eventID),
	}), nil
}

func getTemplatesHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	templates, err := client.GetTemplates()
	if err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("获取模板失败: %v", err)), nil
	}

	return mcp.NewToolResult(templates), nil
}

func linkTemplateHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	hostID := getStringParam(params, "host_id")
	templateID := getStringParam(params, "template_id")

	if hostID == "" || templateID == "" {
		return mcp.NewErrorResult("主机ID和模板ID不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	if err := client.LinkTemplate(hostID, templateID); err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("关联模板失败: %v", err)), nil
	}

	return mcp.NewToolResult(map[string]interface{}{
		"message": fmt.Sprintf("模板 %s 关联到主机 %s 成功", templateID, hostID),
	}), nil
}

func unlinkTemplateHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	hostID := getStringParam(params, "host_id")
	templateID := getStringParam(params, "template_id")

	if hostID == "" || templateID == "" {
		return mcp.NewErrorResult("主机ID和模板ID不能为空"), nil
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return mcp.NewErrorResult("未找到指定的实例"), nil
	}

	if err := client.UnlinkTemplate(hostID, templateID); err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("移除模板失败: %v", err)), nil
	}

	return mcp.NewToolResult(map[string]interface{}{
		"message": fmt.Sprintf("模板 %s 从主机 %s 移除成功", templateID, hostID),
	}), nil
}

func listInstancesHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instances := pool.ListInstances()
	return mcp.NewToolResult(instances), nil
}

func switchInstanceHandler(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	instanceName := getStringParam(params, "instance")
	if instanceName == "" {
		return mcp.NewErrorResult("实例名称不能为空"), nil
	}

	if err := pool.SetDefault(instanceName); err != nil {
		return mcp.NewErrorResult(fmt.Sprintf("切换实例失败: %v", err)), nil
	}

	return mcp.NewToolResult(map[string]interface{}{
		"message": fmt.Sprintf("已切换到实例: %s", instanceName),
	}), nil
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
