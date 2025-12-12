package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetItemsHandler 获取监控项列表
func GetItemsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用GetItemsHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	hostID := ""
	itemName := ""
	itemType := ""
	page := 1
	pageSize := 20

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["item_name"].(string); ok {
		itemName = v
	}
	if v, ok := args["item_type"].(string); ok {
		itemType = v
	}
	if v, ok := args["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok && v > 0 && v <= 100 {
		pageSize = int(v)
	}

	GetSugar().Infof("获取监控项列表 - 实例: %s, 主机ID: %s, 监控项名称: %s, 类型: %s, 页码: %d, 每页数量: %d",
		instanceName, hostID, itemName, itemType, page, pageSize)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	// 获取所有监控项，然后手动分页
	allItems, err := client.GetItems(hostID, itemName)
	if err != nil {
		GetSugar().Errorf("获取监控项列表失败: %v", err)
		return nil, fmt.Errorf("获取监控项列表失败: %v", err)
	}

	total := len(allItems)
	totalPages := (total + pageSize - 1) / pageSize

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	var items []map[string]interface{}
	if start < total {
		items = allItems[start:end]
	}

	GetSugar().Infof("成功获取监控项列表，共 %d 个监控项，当前第 %d 页，共 %d 页", len(items), page, totalPages)

	// 构建分页响应结果
	result := map[string]interface{}{
		"items": items,
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

// GetItemDataHandler 获取监控项数据
func GetItemDataHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用GetItemDataHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	itemID := ""
	history := 0
	timeRange := "1h"

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["item_id"].(string); ok {
		itemID = v
	}
	if v, ok := args["history"].(float64); ok {
		history = int(v)
	}
	if v, ok := args["time_range"].(string); ok {
		timeRange = v
	}

	if itemID == "" {
		return nil, fmt.Errorf("监控项ID不能为空")
	}

	// 解析时间范围
	timeFrom, timeTill, err := parseTimeRange(timeRange)
	if err != nil {
		GetSugar().Errorf("解析时间范围失败: %v", err)
		return nil, fmt.Errorf("解析时间范围失败: %v", err)
	}

	GetSugar().Infof("获取监控项数据 - 实例: %s, 监控项ID: %s, 历史数据类型: %d, 时间范围: %s (%s 至 %s)",
		instanceName, itemID, history, timeRange, timeFrom, timeTill)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	// 获取监控项信息
	item, err := client.GetItemInfo(itemID)
	if err != nil {
		GetSugar().Errorf("获取监控项信息失败: %v", err)
		return nil, fmt.Errorf("获取监控项信息失败: %v", err)
	}

	// 获取监控项历史数据（使用解析后的时间范围）
	historyData, err := client.GetItemDataWithTimeRange(itemID, history, timeFrom, timeTill)
	if err != nil {
		GetSugar().Errorf("获取监控项历史数据失败: %v", err)
		return nil, fmt.Errorf("获取监控项历史数据失败: %v", err)
	}

	GetSugar().Infof("成功获取监控项 %s 的数据，共 %d 条历史记录", itemID, len(historyData))

	// 计算统计信息
	stats := calculateHistoryStats(historyData)
	// 应用返回限制
	returnData := historyData
	// 构建响应结果
	result := map[string]interface{}{
		"item":    item,
		"history": returnData,
		"count":   len(returnData),
		"total":   len(historyData),
		"stats":   stats,
		"time_range": map[string]string{
			"from": timeFrom,
			"till": timeTill,
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

// CreateItemHandler 创建监控项
func CreateItemHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用CreateItemHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	hostID := ""
	itemName := ""
	key := ""
	itemType := "0"
	valueType := "0"
	delay := "30s"

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["item_name"].(string); ok {
		itemName = v
	}
	if v, ok := args["key"].(string); ok {
		key = v
	}
	if v, ok := args["type"].(string); ok {
		itemType = v
	}
	if v, ok := args["value_type"].(string); ok {
		valueType = v
	}
	if v, ok := args["delay"].(string); ok {
		delay = v
	}

	if hostID == "" || itemName == "" || key == "" {
		return nil, fmt.Errorf("主机ID、监控项名称和键值不能为空")
	}

	GetSugar().Infof("创建监控项 - 实例: %s, 主机ID: %s, 监控项名称: %s, 键值: %s, 类型: %s, 值类型: %s",
		instanceName, hostID, itemName, key, itemType, valueType)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	itemID, err := client.CreateItem(hostID, itemName, key, itemType, valueType, delay)
	if err != nil {
		GetSugar().Errorf("创建监控项失败: %v", err)
		return nil, fmt.Errorf("创建监控项失败: %v", err)
	}

	GetSugar().Infof("监控项创建成功，ID: %s", itemID)

	resultData, _ := json.Marshal(map[string]interface{}{
		"itemid":  itemID,
		"message": fmt.Sprintf("监控项 %s 创建成功", itemName),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

// parseTimeRange 解析时间范围字符串，返回开始和结束时间
func parseTimeRange(timeRange string) (string, string, error) {
	// 使用正则表达式解析时间范围格式
	re := regexp.MustCompile(`^(\d+)([wdhms])$`)
	matches := re.FindStringSubmatch(timeRange)

	if len(matches) != 3 {
		return "", "", fmt.Errorf("无效的时间范围格式: %s，支持格式如: 1w, 7d, 3d, 2h, 10m", timeRange)
	}

	amount, err := strconv.Atoi(matches[1])
	if err != nil {
		return "", "", fmt.Errorf("解析时间数量失败: %v", err)
	}

	unit := matches[2]
	var duration time.Duration

	switch unit {
	case "w":
		duration = time.Duration(amount) * 7 * 24 * time.Hour
	case "d":
		duration = time.Duration(amount) * 24 * time.Hour
	case "h":
		duration = time.Duration(amount) * time.Hour
	case "m":
		duration = time.Duration(amount) * time.Minute
	case "s":
		duration = time.Duration(amount) * time.Second
	default:
		return "", "", fmt.Errorf("不支持的时间单位: %s", unit)
	}

	// 计算时间范围
	now := time.Now()
	startTime := now.Add(-duration)

	// 格式化为 Zabbix API 期望的格式 (YYYY-MM-DD HH:MM:SS)
	timeFrom := startTime.Format("2006-01-02 15:04:05")
	timeTill := now.Format("2006-01-02 15:04:05")

	return timeFrom, timeTill, nil
}

// calculateHistoryStats 计算历史数据的统计信息
func calculateHistoryStats(historyData []map[string]interface{}) map[string]interface{} {
	if len(historyData) == 0 {
		return map[string]interface{}{
			"min":   0,
			"max":   0,
			"avg":   0,
			"count": 0,
		}
	}

	var sum, min, max float64
	var count int
	var firstValue = true

	for _, data := range historyData {
		if valueStr, ok := data["value"].(string); ok {
			if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
				if firstValue {
					min, max = value, value
					firstValue = false
				} else {
					if value < min {
						min = value
					}
					if value > max {
						max = value
					}
				}
				sum += value
				count++
			}
		}
	}

	avg := 0.0
	if count > 0 {
		avg = sum / float64(count)
	}

	return map[string]interface{}{
		"min":   math.Round(min*100) / 100,
		"max":   math.Round(max*100) / 100,
		"avg":   math.Round(avg*100) / 100,
		"count": count,
	}
}
