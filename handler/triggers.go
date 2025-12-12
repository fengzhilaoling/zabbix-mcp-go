package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetTriggersHandler 获取触发器列表
func GetTriggersHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用GetTriggersHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	hostID := ""
	triggerName := ""
	activeOnly := true
	page := 1
	pageSize := 20

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["trigger_name"].(string); ok {
		triggerName = v
	}
	if v, ok := args["active_only"].(bool); ok {
		activeOnly = v
	}
	if v, ok := args["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok && v > 0 && v <= 100 {
		pageSize = int(v)
	}

	GetSugar().Infof("获取触发器列表 - 实例: %s, 主机ID: %s, 触发器名称: %s, 仅活跃: %t, 页码: %d, 每页数量: %d",
		instanceName, hostID, triggerName, activeOnly, page, pageSize)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	// 获取所有触发器，然后手动分页
	allTriggers, err := client.GetTriggers(hostID, activeOnly)
	if err != nil {
		GetSugar().Errorf("获取触发器列表失败: %v", err)
		return nil, fmt.Errorf("获取触发器列表失败: %v", err)
	}

	// 根据触发器名称过滤
	var filteredTriggers []map[string]interface{}
	if triggerName != "" {
		for _, trigger := range allTriggers {
			if name, ok := trigger["description"].(string); ok && strings.Contains(name, triggerName) {
				filteredTriggers = append(filteredTriggers, trigger)
			}
		}
	} else {
		filteredTriggers = allTriggers
	}

	total := len(filteredTriggers)
	totalPages := (total + pageSize - 1) / pageSize

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	var triggers []map[string]interface{}
	if start < total {
		triggers = filteredTriggers[start:end]
	}

	GetSugar().Infof("成功获取触发器列表，共 %d 个触发器，当前第 %d 页，共 %d 页", len(triggers), page, totalPages)

	// 构建分页响应结果
	result := map[string]interface{}{
		"triggers": triggers,
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

// GetTriggerEventsHandler 获取触发器事件
func GetTriggerEventsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用GetTriggerEventsHandler，参数: %+v", req.Params.Arguments)

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
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 1000 {
		limit = int(v)
	}

	if triggerID == "" {
		return nil, fmt.Errorf("触发器ID不能为空")
	}

	GetSugar().Infof("获取触发器事件 - 实例: %s, 触发器ID: %s, 限制: %d", instanceName, triggerID, limit)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	events, err := client.GetTriggerEvents(triggerID, limit)
	if err != nil {
		GetSugar().Errorf("获取触发器事件失败: %v", err)
		return nil, fmt.Errorf("获取触发器事件失败: %v", err)
	}

	GetSugar().Infof("成功获取触发器 %s 的事件，共 %d 条记录", triggerID, len(events))

	// 构建响应结果
	result := map[string]interface{}{
		"trigger_id": triggerID,
		"events":     events,
		"count":      len(events),
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

// AcknowledgeEventHandler 确认事件
func AcknowledgeEventHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用AcknowledgeEventHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	eventIDs := []string{}
	message := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["event_ids"].([]interface{}); ok {
		for _, id := range v {
			if strID, ok := id.(string); ok {
				eventIDs = append(eventIDs, strID)
			}
		}
	}
	if v, ok := args["message"].(string); ok {
		message = v
	}

	if len(eventIDs) == 0 {
		return nil, fmt.Errorf("事件ID列表不能为空")
	}

	GetSugar().Infof("确认事件 - 实例: %s, 事件ID: %v, 消息: %s", instanceName, eventIDs, message)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	err := client.MassAcknowledgeEvents(eventIDs, message)
	if err != nil {
		GetSugar().Errorf("确认事件失败: %v", err)
		return nil, fmt.Errorf("确认事件失败: %v", err)
	}

	GetSugar().Infof("成功确认 %d 个事件", len(eventIDs))

	resultData, _ := json.Marshal(map[string]interface{}{
		"event_ids": eventIDs,
		"message":   fmt.Sprintf("成功确认 %d 个事件", len(eventIDs)),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}
