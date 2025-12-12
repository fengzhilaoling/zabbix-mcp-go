package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListInstancesHandler 列出所有实例
func ListInstancesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Info("调用ListInstancesHandler")

	instances := pool.ListInstances()
	GetSugar().Infof("成功获取实例列表，共 %d 个实例", len(instances))

	// 提取实例名称列表
	var instanceNames []string
	for _, instance := range instances {
		if name, ok := instance["name"].(string); ok {
			instanceNames = append(instanceNames, name)
		}
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"instances": instanceNames,
		"count":     len(instanceNames),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

// SwitchInstanceHandler 切换当前实例
func SwitchInstanceHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用SwitchInstanceHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}

	if instanceName == "" {
		return nil, fmt.Errorf("实例名称不能为空")
	}

	GetSugar().Infof("切换实例 - 目标实例: %s", instanceName)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例: %s", instanceName)
	}
	_ = getZabbixClient(clientRaw)

	// 设置默认实例
	if err := pool.SetDefault(instanceName); err != nil {
		GetSugar().Errorf("设置默认实例失败: %v", err)
		return nil, fmt.Errorf("设置默认实例失败: %v", err)
	}

	GetSugar().Infof("成功切换到实例: %s", instanceName)

	resultData, _ := json.Marshal(map[string]interface{}{
		"current_instance": instanceName,
		"message":          fmt.Sprintf("成功切换到实例: %s", instanceName),
	})
	return mcp.NewToolResultText(string(resultData)), nil
}

// GetInstancesInfoHandler 获取实例信息
func GetInstancesInfoHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	GetSugar().Infof("调用GetInstancesInfoHandler，参数: %+v", req.Params.Arguments)

	args := req.Params.Arguments
	instanceName := ""

	if v, ok := args["instance"].(string); ok {
		instanceName = v
	}

	GetSugar().Infof("获取实例信息 - 实例: %s", instanceName)

	clientRaw := pool.GetClient(instanceName)
	if clientRaw == nil {
		GetSugar().Errorf("未找到指定的实例: %s", instanceName)
		return nil, fmt.Errorf("未找到指定的实例: %s", instanceName)
	}
	// 只检查客户端是否存在，不调用具体方法

	GetSugar().Infof("成功获取实例 %s 的信息", instanceName)

	result := map[string]interface{}{
		"instance": instanceName,
		"status":   "connected",
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
