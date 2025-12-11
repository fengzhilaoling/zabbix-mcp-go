package zabbix

import (
	"fmt"
)

// CallWithVersion 版本感知的API调用
func (c *ZabbixClient) CallWithVersion(method string, params interface{}) (interface{}, error) {
	detector := NewVersionDetector(c)

	// 获取版本信息
	version, err := detector.DetectVersion()
	if err != nil {
		return nil, err
	}

	// 根据版本调整方法和参数
	adaptedMethod := method
	adaptedParams := params

	// 版本特定的适配逻辑
	switch method {
	case "problem.get":
		if version.Major < 4 {
			// 旧版本使用 trigger.get
			adaptedMethod = "trigger.get"
			// 转换参数
			if paramMap, ok := params.(map[string]interface{}); ok {
				adaptedParams = convertProblemToTriggerParams(paramMap)
			}
		}
	case "sla.get":
		if version.Major < 5 {
			return nil, fmt.Errorf("SLA功能在Zabbix %s版本中不支持", version.String())
		}
	}

	return c.Call(adaptedMethod, adaptedParams)
}

// CallWithFallback 带回退机制的API调用
func (c *ZabbixClient) CallWithFallback(primaryMethod, fallbackMethod string, params interface{}) (interface{}, error) {
	// 首先尝试主要方法
	result, err := c.Call(primaryMethod, params)
	if err != nil {
		// 检查是否是方法不存在的错误
		if rpcErr, ok := err.(*RPCError); ok && rpcErr.Code == -32601 {
			// 尝试回退方法
			if fallbackMethod != "" {
				// 使用日志记录
				fmt.Printf("主要方法 %s 失败，尝试回退方法 %s\n", primaryMethod, fallbackMethod)
				return c.Call(fallbackMethod, params)
			}
		}
		return nil, err
	}
	return result, nil
}

// convertProblemToTriggerParams 将problem.get参数转换为trigger.get参数
func convertProblemToTriggerParams(problemParams map[string]interface{}) map[string]interface{} {
	triggerParams := make(map[string]interface{})

	// 复制通用参数
	for key, value := range problemParams {
		switch key {
		case "eventids", "groupids", "hostids", "objectids", "applicationids":
			triggerParams[key] = value
		case "limit", "output":
			triggerParams[key] = value
		}
	}

	// 添加触发器特定的过滤条件
	triggerParams["filter"] = map[string]interface{}{
		"value": 1, // 只获取问题状态的触发器
	}

	return triggerParams
}

// VersionCompatibleCall 根据版本兼容性调用API
func (c *ZabbixClient) VersionCompatibleCall(method string, params interface{}, minVersion string) (interface{}, error) {
	detector := NewVersionDetector(c)

	// 获取版本信息
	version, err := detector.DetectVersion()
	if err != nil {
		return nil, err
	}

	// 解析最低版本要求
	minVer, err := detector.parseVersion(minVersion)
	if err != nil {
		return nil, fmt.Errorf("解析最低版本失败: %w", err)
	}

	// 检查版本兼容性
	if !isVersionCompatible(version, minVer) {
		return nil, fmt.Errorf("当前Zabbix版本 %s 不满足最低版本要求 %s", version.String(), minVersion)
	}

	return c.Call(method, params)
}

// isVersionCompatible 检查版本兼容性
func isVersionCompatible(current, minimum *VersionInfo) bool {
	if current.Major > minimum.Major {
		return true
	}
	if current.Major < minimum.Major {
		return false
	}
	// Major版本相同，比较Minor版本
	if current.Minor > minimum.Minor {
		return true
	}
	if current.Minor < minimum.Minor {
		return false
	}
	// Major和Minor版本相同，比较Patch版本
	return current.Patch >= minimum.Patch
}
