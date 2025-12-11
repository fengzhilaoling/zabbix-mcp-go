package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// 工具处理器示例
func getHostsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用getHostsHandler，参数: %+v", req.Params.Arguments)

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

	GetSugar().Infof("获取主机列表 - 实例: %s, 组ID: %s, 主机名: %s", instanceName, groupID, hostName)

	client := pool.GetClient(instanceName)
	if client == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}

	hosts, err := client.GetHosts(groupID, hostName)
	if err != nil {
		GetSugar().Errorf("获取主机列表失败: %v", err)
		return nil, fmt.Errorf("获取主机列表失败: %v", err)
	}

	GetSugar().Infof("成功获取主机列表，共 %d 台主机", len(hosts))

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
	GetSugar().Infof("调用getHostByNameHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	hostName := ""
	detailed := false

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_name"].(string); ok {
		hostName = v
	}
	if v, ok := args["detailed"].(bool); ok {
		detailed = v
	}

	if hostName == "" {
		return nil, fmt.Errorf("主机名不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	// 根据detailed参数决定使用哪种查询方式
	var host map[string]interface{}
	var err error
	if detailed {
		GetSugar().Infof("使用详细模式获取主机信息: %s", hostName)
		host, err = client.GetHostByName(hostName)
	} else {
		GetSugar().Infof("使用轻量级模式获取主机信息: %s", hostName)
		host, err = client.GetHostByNameLite(hostName)
	}

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
	GetSugar().Infof("调用createHostHandler，参数: %+v", req.Params.Arguments)

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
	GetSugar().Infof("调用deleteHostHandler，参数: %+v", req.Params.Arguments)

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
	GetSugar().Infof("调用getItemsHandler，参数: %+v", req.Params.Arguments)

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

// getInstancesInfoHandler 获取所有实例详细信息
func getInstancesInfoHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Info("调用getInstancesInfoHandler")

	if pool == nil {
		return mcp.NewToolResultError("Zabbix连接池未初始化"), nil
	}

	// 获取所有实例信息
	instancesInfo := pool.GetAllInstancesInfo()

	// 记录详细日志
	if instances, ok := instancesInfo["instances"].([]map[string]interface{}); ok {
		for _, instance := range instances {
			GetSugar().Infof("实例信息 - 名称: %s, 地址: %s, 认证方式: %s, 状态: %s, 版本: %s, 默认实例: %v",
				instance["name"], instance["url"], instance["auth_type"],
				instance["status"], instance["version"], instance["is_default"])
		}
	}

	resultData, _ := json.Marshal(instancesInfo)
	return mcp.NewToolResultText(string(resultData)), nil
}

// getHostTemplatesHandler 获取指定主机的模板信息
func getHostTemplatesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用getHostTemplatesHandler，参数: %+v", req.Params.Arguments)

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

	GetSugar().Infof("获取主机模板 - 实例: %s, 主机ID: %s", instanceName, hostID)

	client := pool.GetClient(instanceName)
	if client == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}

	templates, err := client.GetTemplatesByHost(hostID)
	if err != nil {
		GetSugar().Errorf("获取主机关联模板失败: %v", err)
		return nil, fmt.Errorf("获取主机关联模板失败: %v", err)
	}

	GetSugar().Infof("成功获取主机 %s 的模板列表，共 %d 个模板", hostID, len(templates))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", templates),
			},
		},
	}, nil
}
