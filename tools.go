package main

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
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
