package handler

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 注册所有工具处理函数
func RegisterTools(s *server.MCPServer) {
	// 主机相关工具
	s.AddTool(
		// 获取主机列表
		mcp.NewTool("get_hosts",
			mcp.WithDescription("获取Zabbix主机列表，支持分页查询"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("group_id", mcp.Description("主机组ID")),
			mcp.WithString("host_name", mcp.Description("主机名模糊筛选，支持通配符*和?")),
			mcp.WithNumber("page", mcp.DefaultNumber(1), mcp.Description("页码，默认为1")),
			mcp.WithNumber("page_size", mcp.DefaultNumber(20), mcp.Description("每页数量，默认为20，最大100")),
		),
		GetHostsHandler,
	)
	// 主机名获取主机信息
	s.AddTool(
		mcp.NewTool("get_host_by_name",
			mcp.WithDescription("根据主机名获取主机信息"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_name", mcp.Required(), mcp.Description("主机名")),
		),
		GetHostByNameHandler,
	)
	// TODO 创建主机 测试
	s.AddTool(
		mcp.NewTool("create_host",
			mcp.WithDescription("创建Zabbix主机"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_name", mcp.Required(), mcp.Description("主机名")),
			mcp.WithString("group_id", mcp.Required(), mcp.Description("主机组ID")),
			mcp.WithString("interface_ip", mcp.Required(), mcp.Description("接口IP地址")),
		),
		CreateHostHandler,
	)
	// TODO 删除主机 测试
	s.AddTool(
		mcp.NewTool("delete_host",
			mcp.WithDescription("删除Zabbix主机"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
		),
		DeleteHostHandler,
	)

	// 监控项相关工具
	s.AddTool(
		// 获取主机监控项 完成
		mcp.NewTool("get_host_items",
			mcp.WithDescription("获取主机监控项，支持监控项名称模糊匹配"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
			mcp.WithString("item_name", mcp.Description("监控项名称模糊匹配，不传入则获取所有监控项")),
		),
		GetItemsHandler,
	)
	// 获取主机监控项数据支持时间范围
	s.AddTool(
		mcp.NewTool("get_item_data",
			mcp.WithDescription("获取监控项数据，通过监控项ID获取数据"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("item_id", mcp.Required(), mcp.Description("监控项ID")),
			mcp.WithString("time_range", mcp.DefaultString("1h"), mcp.Description("时间范围，支持格式：1w(1周), 7d(7天), 3d(3天), 2h(2小时), 10m(10分钟)")),
			mcp.WithNumber("history", mcp.DefaultNumber(0), mcp.Description("历史数据类型")),
		),
		GetItemDataHandler,
	)
	// TODO 创建监控项 测试
	s.AddTool(
		mcp.NewTool("create_item",
			mcp.WithDescription("创建监控项"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
			mcp.WithString("item_name", mcp.Required(), mcp.Description("监控项名称")),
			mcp.WithString("key", mcp.Required(), mcp.Description("监控项键值")),
			mcp.WithString("type", mcp.Description("监控项类型")),
			mcp.WithString("value_type", mcp.Description("值类型")),
			mcp.WithString("delay", mcp.Description("更新间隔")),
		),
		CreateItemHandler,
	)

	// TODO 触发器相关工具  测试
	s.AddTool(
		mcp.NewTool("get_triggers",
			mcp.WithDescription("获取触发器列表"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Description("主机ID")),
			mcp.WithBoolean("active", mcp.Description("只显示启用的触发器")),
		),
		GetTriggersHandler,
	)
	s.AddTool(
		mcp.NewTool("get_trigger_events",
			mcp.WithDescription("获取触发器事件"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("trigger_id", mcp.Required(), mcp.Description("触发器ID")),
			mcp.WithNumber("limit", mcp.DefaultNumber(100), mcp.Description("事件数量限制")),
		),
		GetTriggerEventsHandler,
	)
	s.AddTool(
		mcp.NewTool("acknowledge_event",
			mcp.WithDescription("确认事件"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("event_id", mcp.Required(), mcp.Description("事件ID")),
			mcp.WithString("message", mcp.Description("确认消息")),
		),
		AcknowledgeEventHandler,
	)

	// 模板相关工具
	s.AddTool(
		mcp.NewTool("get_templates",
			mcp.WithDescription("获取模板列表"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
		),
		GetTemplatesHandler,
	)
	// info 获取模板信息 完成
	s.AddTool(
		mcp.NewTool("get_host_templates",
			mcp.WithDescription("获取指定主机的模板信息"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
		),
		GetHostTemplatesHandler,
	)
	s.AddTool(
		mcp.NewTool("link_template",
			mcp.WithDescription("关联模板到主机"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
			mcp.WithString("template_id", mcp.Required(), mcp.Description("模板ID")),
		),
		LinkTemplateHandler,
	)
	s.AddTool(
		mcp.NewTool("unlink_template",
			mcp.WithDescription("从主机移除模板"),
			mcp.WithString("instance", mcp.Description("Zabbix实例名称")),
			mcp.WithString("host_id", mcp.Required(), mcp.Description("主机ID")),
			mcp.WithString("template_id", mcp.Required(), mcp.Description("模板ID")),
		),
		UnlinkTemplateHandler,
	)

	// info 多实例管理 完成
	s.AddTool(
		mcp.NewTool("list_instances",
			mcp.WithDescription("列出所有Zabbix实例"),
		),
		ListInstancesHandler,
	)
	s.AddTool(
		mcp.NewTool("switch_instance",
			mcp.WithDescription("切换当前使用的实例"),
			mcp.WithString("instance", mcp.Required(), mcp.Description("实例名称")),
		),
		SwitchInstanceHandler,
	)

	// info 实例信息工具 完成
	s.AddTool(
		mcp.NewTool("get_instances_info",
			mcp.WithDescription("获取所有Zabbix实例的详细信息"),
		),
		GetInstancesInfoHandler,
	)
}
