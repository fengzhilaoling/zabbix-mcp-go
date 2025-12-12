package zabbix

import (
	"fmt"
	"strconv"
	"time"
)

// GetItems 获取主机监控项，支持监控项名称模糊匹配
func (c *ZabbixClient) GetItems(hostID string, itemNameFilter string) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":             "extend",
		"hostids":            hostID,
		"selectApplications": "extend",
		"selectTriggers":     "extend",
		"preservekeys":       1, // 保持关联数组格式，便于验证
	}

	// 如果提供了监控项名称过滤条件，添加模糊匹配
	if itemNameFilter != "" {
		// 添加通配符以实现真正的模糊匹配
		searchPattern := fmt.Sprintf("*%s*", itemNameFilter)
		params["search"] = map[string]interface{}{
			"name": searchPattern,
			"key_": searchPattern, // 同时搜索监控项键值
		}
		params["searchWildcardsEnabled"] = 1 // 启用通配符搜索
		params["searchByAny"] = 1            // 匹配任意字段即可
	}

	result, err := c.Call("item.get", params)
	if err != nil {
		return nil, fmt.Errorf("API调用失败: %v", err)
	}

	// 检查响应格式
	if result == nil {
		return []map[string]interface{}{}, nil // 返回空数组而不是错误
	}

	items, ok := result.([]interface{})
	if !ok {
		// 尝试其他可能的格式
		if itemMap, ok := result.(map[string]interface{}); ok {
			// 如果是map格式，转换为数组
			var itemList []map[string]interface{}
			for _, item := range itemMap {
				if itemData, ok := item.(map[string]interface{}); ok {
					if itemHostID, exists := itemData["hostid"].(string); exists && itemHostID == hostID {
						itemList = append(itemList, itemData)
					}
				}
			}
			return itemList, nil
		}
		return nil, fmt.Errorf("响应格式错误: 期望数组，得到 %T", result)
	}

	var itemList []map[string]interface{}
	for _, i := range items {
		if item, ok := i.(map[string]interface{}); ok {
			// 双重验证：确保监控项确实属于指定主机
			if itemHostID, exists := item["hostid"].(string); exists && itemHostID == hostID {
				itemList = append(itemList, item)
			}
		}
	}

	return itemList, nil
}

// GetItemData 获取监控项数据
func (c *ZabbixClient) GetItemData(itemID string, history, limit int) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":    "extend",
		"history":   history,
		"itemids":   itemID,
		"sortfield": "clock",
		"sortorder": "DESC",
		"limit":     limit,
	}

	// 修复：移除了无意义的条件判断
	method := "history.get"

	result, err := c.Call(method, params)
	if err != nil {
		return nil, err
	}

	data, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var dataList []map[string]interface{}
	for _, d := range data {
		if item, ok := d.(map[string]interface{}); ok {
			dataList = append(dataList, item)
		}
	}

	return dataList, nil
}

// GetItemDataWithTimeRange 获取监控项数据（带时间范围）
func (c *ZabbixClient) GetItemDataWithTimeRange(itemID string, history int, timeFrom, timeTill string) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":    "extend",
		"history":   history,
		"itemids":   itemID,
		"sortfield": "clock",
		"sortorder": "DESC",
	}

	// 添加时间范围参数 - 将时间字符串转换为 Unix 时间戳
	if timeFrom != "" {
		// 尝试解析时间字符串并转换为 Unix 时间戳
		if t, err := time.Parse("2006-01-02 15:04:05", timeFrom); err == nil {
			params["time_from"] = t.Unix()
		} else {
			// 如果解析失败，尝试直接作为 Unix 时间戳字符串
			if timestamp, err := strconv.ParseInt(timeFrom, 10, 64); err == nil {
				params["time_from"] = timestamp
			}
		}
	}
	if timeTill != "" {
		// 尝试解析时间字符串并转换为 Unix 时间戳
		if t, err := time.Parse("2006-01-02 15:04:05", timeTill); err == nil {
			params["time_till"] = t.Unix()
		} else {
			// 如果解析失败，尝试直接作为 Unix 时间戳字符串
			if timestamp, err := strconv.ParseInt(timeTill, 10, 64); err == nil {
				params["time_till"] = timestamp
			}
		}
	}

	method := "history.get"

	result, err := c.Call(method, params)
	if err != nil {
		return nil, err
	}

	data, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var dataList []map[string]interface{}
	for _, d := range data {
		if item, ok := d.(map[string]interface{}); ok {
			dataList = append(dataList, item)
		}
	}

	return dataList, nil
}

// CreateItem 创建监控项
func (c *ZabbixClient) CreateItem(hostID, name, key, type_, valueType, delay string) (string, error) {
	if type_ == "" {
		type_ = "0" // Zabbix agent
	}
	if valueType == "" {
		valueType = "3" // Numeric unsigned
	}
	if delay == "" {
		delay = "60s"
	}

	params := map[string]interface{}{
		"name":        name,
		"key_":        key,
		"hostid":      hostID,
		"type":        type_,
		"value_type":  valueType,
		"delay":       delay,
		"interfaceid": "$1", // 使用默认接口
	}

	result, err := c.Call("item.create", params)
	if err != nil {
		return "", err
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误")
	}

	itemIDs, ok := response["itemids"].([]interface{})
	if !ok || len(itemIDs) == 0 {
		return "", fmt.Errorf("创建监控项失败")
	}

	return itemIDs[0].(string), nil
}

// UpdateItem 更新监控项
func (c *ZabbixClient) UpdateItem(itemID string, params map[string]interface{}) error {
	params["itemid"] = itemID
	_, err := c.Call("item.update", params)
	return err
}

// DeleteItem 删除监控项
func (c *ZabbixClient) DeleteItem(itemID string) error {
	params := []string{itemID}
	_, err := c.Call("item.delete", params)
	return err
}

// GetItemByKey 根据监控项键获取监控项信息
func (c *ZabbixClient) GetItemByKey(hostID, key string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":  "extend",
		"hostids": hostID,
		"filter": map[string]string{
			"key_": key,
		},
	}

	result, err := c.Call("item.get", params)
	if err != nil {
		return nil, err
	}

	items, ok := result.([]interface{})
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("监控项不存在")
	}

	if item, ok := items[0].(map[string]interface{}); ok {
		return item, nil
	}

	return nil, fmt.Errorf("响应格式错误")
}

// GetItemInfo 根据监控项ID获取监控项详细信息
func (c *ZabbixClient) GetItemInfo(itemID string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":  "extend",
		"itemids": itemID,
	}

	result, err := c.Call("item.get", params)
	if err != nil {
		return nil, err
	}

	items, ok := result.([]interface{})
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("监控项不存在")
	}

	if item, ok := items[0].(map[string]interface{}); ok {
		return item, nil
	}

	return nil, fmt.Errorf("响应格式错误")
}
