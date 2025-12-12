package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetTemplatesHandler 获取模板列表
func GetTemplatesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用GetTemplatesHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	templateName := ""
	page := 1
	pageSize := 20

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["template_name"].(string); ok {
		templateName = v
	}
	if v, ok := args["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok && v > 0 && v <= 100 {
		pageSize = int(v)
	}

	GetSugar().Infof("获取模板列表 - 实例: %s, 模板名称: %s, 页码: %d, 每页数量: %d",
		instanceName, templateName, page, pageSize)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	// 获取所有模板，然后手动分页和过滤
	allTemplates, err := client.GetTemplates()
	if err != nil {
		GetSugar().Errorf("获取模板列表失败: %v", err)
		return nil, fmt.Errorf("获取模板列表失败: %v", err)
	}

	// 根据模板名称过滤
	var filteredTemplates []map[string]interface{}
	if templateName != "" {
		for _, template := range allTemplates {
			if name, ok := template["name"].(string); ok && strings.Contains(name, templateName) {
				filteredTemplates = append(filteredTemplates, template)
			}
		}
	} else {
		filteredTemplates = allTemplates
	}

	total := len(filteredTemplates)
	totalPages := (total + pageSize - 1) / pageSize

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	var templates []map[string]interface{}
	if start < total {
		templates = filteredTemplates[start:end]
	}

	GetSugar().Infof("成功获取模板列表，共 %d 个模板，当前第 %d 页，共 %d 页", len(templates), page, totalPages)

	// 构建分页响应结果
	result := map[string]interface{}{
		"templates": templates,
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

// LinkTemplateHandler 关联模板到主机
func LinkTemplateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用LinkTemplateHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	hostID := ""
	templateIDs := []string{}

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["template_ids"].([]interface{}); ok {
		for _, id := range v {
			if strID, ok := id.(string); ok {
				templateIDs = append(templateIDs, strID)
			}
		}
	}

	if hostID == "" || len(templateIDs) == 0 {
		return nil, fmt.Errorf("主机ID和模板ID列表不能为空")
	}

	GetSugar().Infof("关联模板 - 实例: %s, 主机ID: %s, 模板ID: %v", instanceName, hostID, templateIDs)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	err := client.LinkTemplates(hostID, templateIDs)
	if err != nil {
		GetSugar().Errorf("关联模板失败: %v", err)
		return nil, fmt.Errorf("关联模板失败: %v", err)
	}

	GetSugar().Infof("成功关联 %d 个模板到主机 %s", len(templateIDs), hostID)

	resultData, _ := json.Marshal(map[string]interface{}{
		"host_id":      hostID,
		"template_ids": templateIDs,
		"message":      fmt.Sprintf("成功关联 %d 个模板到主机", len(templateIDs)),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

// UnlinkTemplateHandler 从主机取消关联模板
func UnlinkTemplateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用UnlinkTemplateHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""
	hostID := ""
	templateIDs := []string{}
	clear := false

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}
	if v, ok := args["host_id"].(string); ok {
		hostID = v
	}
	if v, ok := args["template_ids"].([]interface{}); ok {
		for _, id := range v {
			if strID, ok := id.(string); ok {
				templateIDs = append(templateIDs, strID)
			}
		}
	}
	if v, ok := args["clear"].(bool); ok {
		clear = v
	}

	if hostID == "" || len(templateIDs) == 0 {
		return nil, fmt.Errorf("主机ID和模板ID列表不能为空")
	}

	GetSugar().Infof("取消关联模板 - 实例: %s, 主机ID: %s, 模板ID: %v, 清除实例: %t", instanceName, hostID, templateIDs, clear)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	err := client.UnlinkTemplates(hostID, templateIDs, clear)
	if err != nil {
		GetSugar().Errorf("取消关联模板失败: %v", err)
		return nil, fmt.Errorf("取消关联模板失败: %v", err)
	}

	GetSugar().Infof("成功从主机 %s 取消关联 %d 个模板", hostID, len(templateIDs))

	resultData, _ := json.Marshal(map[string]interface{}{
		"host_id":      hostID,
		"template_ids": templateIDs,
		"clear":        clear,
		"message":      fmt.Sprintf("成功取消关联 %d 个模板", len(templateIDs)),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}
