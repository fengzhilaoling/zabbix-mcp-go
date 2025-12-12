package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetHostsHandler 获取主机列表
func GetHostsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用GetHostsHandler，参数: %+v", req.Params.Arguments)

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

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	// 获取所有主机，然后手动分页
	allHosts, err := client.GetHosts(groupID, hostName)
	if err != nil {
		GetSugar().Errorf("获取主机列表失败: %v", err)
		return nil, fmt.Errorf("获取主机列表失败: %v", err)
	}

	total := len(allHosts)
	totalPages := (total + pageSize - 1) / pageSize

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	var hosts []map[string]interface{}
	if start < total {
		hosts = allHosts[start:end]
	}

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

// GetHostByNameHandler 根据主机名获取主机信息
func GetHostByNameHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用GetHostByNameHandler，参数: %+v", req.Params.Arguments)

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

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

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

// CreateHostHandler 创建主机
func CreateHostHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用CreateHostHandler，参数: %+v", req.Params.Arguments)

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

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

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

// DeleteHostHandler 删除主机
func DeleteHostHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用DeleteHostHandler，参数: %+v", req.Params.Arguments)

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

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

	err := client.DeleteHost(hostID)
	if err != nil {
		return nil, fmt.Errorf("删除主机失败: %v", err)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"hostid":  hostID,
		"message": "主机删除成功",
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

// GetHostTemplatesHandler 获取主机关联的模板
func GetHostTemplatesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用GetHostTemplatesHandler，参数: %+v", req.Params.Arguments)

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

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例")
	}
	client := getZabbixClient(clientRaw)

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
	// GetSugar().Infof("返回模板信息: %v", MustJSON(templateInfo))
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
