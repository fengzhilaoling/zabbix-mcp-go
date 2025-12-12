package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// 工具处理器示例
func getHostsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用getHostsHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	groupID := ""
	hostName := ""
	page := 1
	pageSize := 20

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["group_id"].(string); ok {
		groupID = v
	}
	if v, ok := args["host_name"].(string); ok {
		hostName = v
	}
	if v, ok := args["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok && v > 0 && v <= 100 {
		pageSize = int(v)
	}

	GetSugar().Infof("获取主机列表 - 实例: %s, 组ID: %s, 主机名: %s, 页码: %d, 每页数量: %d", instanceName, groupID, hostName, page, pageSize)

	client := pool.GetClient(instanceName)
	if client == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}

	hosts, total, err := client.GetHostsWithPagination(groupID, hostName, page, pageSize)
	if err != nil {
		GetSugar().Errorf("获取主机列表失败: %v", err)
		return nil, fmt.Errorf("获取主机列表失败: %v", err)
	}

	totalPages := (total + pageSize - 1) / pageSize
	GetSugar().Infof("成功获取主机列表，共 %d 台主机，当前第 %d 页，共 %d 页", len(hosts), page, totalPages)

	// 构建分页响应结果
	result := map[string]interface{}{
		"hosts": hosts,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		GetSugar().Errorf("序列化结果失败: %v", err)
		return nil, fmt.Errorf("序列化结果失败: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(resultJSON),
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
	itemName := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["item_name"].(string); ok {
		itemName = v
	}

	if hostID == "" {
		return nil, fmt.Errorf("主机ID不能为空")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	GetSugar().Infof("开始获取主机 %s 的监控项", hostID)
	items, err := client.GetItems(hostID, itemName)
	if err != nil {
		GetSugar().Errorf("获取监控项失败: %v", err)
		return nil, fmt.Errorf("获取监控项失败: %v", err)
	}

	GetSugar().Infof("从API获取到 %d 个监控项", len(items))

	// 过滤确保只包含指定主机的监控项
	var filteredItems []map[string]interface{}
	for _, item := range items {
		itemHostID, exists := item["hostid"].(string)
		itemDisplayName, _ := item["name"].(string)
		itemKey, _ := item["key_"].(string)

		GetSugar().Debugf("检查监控项 - hostid: %s, name: %s, key: %s, 是否匹配目标hostid: %v",
			itemHostID, itemDisplayName, itemKey, exists && itemHostID == hostID)

		if exists && itemHostID == hostID {
			// 如果指定了监控项名称，进行额外的模糊匹配过滤
			if itemName != "" {
				itemNameLower := strings.ToLower(itemDisplayName)
				itemKeyLower := strings.ToLower(itemKey)
				searchLower := strings.ToLower(itemName)

				// 检查名称或键值是否包含搜索词（不区分大小写）
				if strings.Contains(itemNameLower, searchLower) || strings.Contains(itemKeyLower, searchLower) {
					filteredItems = append(filteredItems, item)
					GetSugar().Debugf("添加监控项到结果: %s (%s)", itemDisplayName, itemKey)
				}
			} else {
				// 没有指定监控项名称，添加所有匹配的监控项
				filteredItems = append(filteredItems, item)
				GetSugar().Debugf("添加监控项到结果: %s (%s)", itemDisplayName, itemKey)
			}
		} else {
			if !exists {
				GetSugar().Warnf("监控项缺少hostid字段: %v", item)
			} else if itemHostID != hostID {
				GetSugar().Warnf("监控项hostid不匹配 - 期望: %s, 实际: %s", hostID, itemHostID)
			}
		}
	}

	GetSugar().Infof("过滤后获取主机 %s 的监控项，共 %d 个", hostID, len(filteredItems))

	// 构建结构化的返回数据
	result := map[string]interface{}{
		"host_id": hostID,
		"count":   len(filteredItems),
		"items":   filteredItems,
	}

	// 记录返回数据到日志
	resultJSON := MustJSON(result)
	// GetSugar().Infof("返回监控项数据: %s", resultJSON)

	return mcp.NewToolResultText(resultJSON), nil
}

func getItemDataHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用getItemDataHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	itemID := ""
	hostName := ""
	itemName := ""
	history := 0
	limit := 100

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["item_id"].(string); ok {
		itemID = v
	}
	if v, ok := args["host_name"].(string); ok {
		hostName = v
	}
	if v, ok := args["item_name"].(string); ok {
		itemName = v
	}
	if v, ok := args["history"].(float64); ok {
		history = int(v)
	}
	if v, ok := args["limit"].(float64); ok {
		limit = int(v)
	}

	// 参数验证：必须提供item_id，或者同时提供host_name和item_name
	if itemID == "" && (hostName == "" || itemName == "") {
		return nil, fmt.Errorf("必须提供监控项ID，或者同时提供主机名和监控项名称")
	}

	client := pool.GetClient(instanceName)
	if client == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}

	// 如果通过主机名+监控项名获取，需要先找到对应的item_id
	if itemID == "" {
		GetSugar().Infof("通过主机名+监控项名获取监控项数据 - 主机: %s, 监控项: %s", hostName, itemName)

		// 先获取主机信息
		hosts, err := client.GetHosts("", hostName)
		if err != nil {
			return nil, fmt.Errorf("获取主机信息失败: %v", err)
		}
		if len(hosts) == 0 {
			return nil, fmt.Errorf("未找到主机: %s", hostName)
		}
		if len(hosts) > 1 {
			return nil, fmt.Errorf("找到多个主机，请提供更精确的主机名")
		}

		hostID := hosts[0]["hostid"].(string)
		host := hosts[0]
		GetSugar().Infof("找到主机: %s (ID: %s)", hostName, hostID)

		// 获取该主机的监控项
		items, err := client.GetItems(hostID, itemName)
		if err != nil {
			return nil, fmt.Errorf("获取监控项失败: %v", err)
		}
		if len(items) == 0 {
			return nil, fmt.Errorf("未找到监控项: %s", itemName)
		}
		if len(items) > 1 {
			return nil, fmt.Errorf("找到多个匹配的监控项，请提供更精确的监控项名称")
		}

		itemID = items[0]["itemid"].(string)
		item := items[0]
		GetSugar().Infof("找到监控项: %s (ID: %s)", itemName, itemID)

		// 获取监控项数据
		data, err := client.GetItemData(itemID, history, limit)
		if err != nil {
			return nil, fmt.Errorf("获取监控项数据失败: %v", err)
		}

		// 构建增强的返回结果，包含主机和监控项信息
		result := map[string]interface{}{
			"host_info": map[string]interface{}{
				"host_id":            hostID,
				"host_name":          host["host"],
				"host_name_readable": host["name"],
			},
			"item_info": map[string]interface{}{
				"item_id":   itemID,
				"item_name": item["name"],
				"item_key":  item["key_"],
			},
			"data_count": len(data),
			"data":       data,
		}

		GetSugar().Infof("成功获取监控项数据 - 主机: %s, 监控项: %s, 数据条数: %d", hostName, itemName, len(data))
		return mcp.NewToolResultText(MustJSON(result)), nil
	}

	// 通过item_id直接获取的情况
	GetSugar().Infof("通过监控项ID获取监控项数据 - ID: %s", itemID)

	// 先获取监控项信息以获取主机信息
	itemInfo, err := client.GetItemInfo(itemID)
	if err != nil {
		GetSugar().Warnf("获取监控项信息失败: %v，将继续获取数据", err)
		// 如果获取监控项信息失败，仍然尝试获取数据
		data, dataErr := client.GetItemData(itemID, history, limit)
		if dataErr != nil {
			return nil, fmt.Errorf("获取监控项数据失败: %v", dataErr)
		}

		// 简化返回结果
		result := map[string]interface{}{
			"item_id":    itemID,
			"data_count": len(data),
			"data":       data,
			"note":       "未获取到监控项详细信息",
		}
		return mcp.NewToolResultText(MustJSON(result)), nil
	}

	// 获取主机信息
	hostID := itemInfo["hostid"].(string)
	hosts, err := client.GetHostByID(hostID)
	if err != nil {
		GetSugar().Warnf("获取主机信息失败: %v", err)
		hosts = []map[string]interface{}{{}}
	}

	host := hosts[0]
	if host == nil {
		host = map[string]interface{}{}
	}

	// 获取监控项数据
	data, err := client.GetItemData(itemID, history, limit)
	if err != nil {
		return nil, fmt.Errorf("获取监控项数据失败: %v", err)
	}

	// 构建增强的返回结果，包含主机和监控项信息
	result := map[string]interface{}{
		"host_info": map[string]interface{}{
			"host_id":            hostID,
			"host_name":          host["host"],
			"host_name_readable": host["name"],
		},
		"item_info": map[string]interface{}{
			"item_id":   itemID,
			"item_name": itemInfo["name"],
			"item_key":  itemInfo["key_"],
		},
		"data_count": len(data),
		"data":       data,
	}

	GetSugar().Infof("成功获取监控项数据 - 主机: %s, 监控项: %s, 数据条数: %d",
		host["host"], itemInfo["name"], len(data))
	return mcp.NewToolResultText(MustJSON(result)), nil
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

	// 格式化模板数据，提取关键信息
	var templateInfo []map[string]interface{}
	for _, template := range templates {
		info := map[string]interface{}{
			"templateid":  template["templateid"],
			"name":        template["name"],
			"host":        template["host"],
			"description": template["description"],
		}
		templateInfo = append(templateInfo, info)
	}
	GetSugar().Infof("返回模板信息: %v", MustJSON(templateInfo))
	// 返回结构化的JSON数据
	result := map[string]interface{}{
		"host_id":   hostID,
		"count":     len(templateInfo),
		"templates": templateInfo,
	}

	resultData, err := json.Marshal(result)
	if err != nil {
		GetSugar().Errorf("JSON序列化失败: %v", err)
		return nil, fmt.Errorf("数据格式化失败: %v", err)
	}

	return mcp.NewToolResultText(string(resultData)), nil
}
