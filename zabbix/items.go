package zabbix

import "fmt"

// GetItems 获取主机监控项
func (c *ZabbixClient) GetItems(hostID string) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":             "extend",
		"hostids":            hostID,
		"selectApplications": "extend",
		"selectTriggers":     "extend",
	}

	result, err := c.Call("item.get", params)
	if err != nil {
		return nil, err
	}

	items, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var itemList []map[string]interface{}
	for _, i := range items {
		if item, ok := i.(map[string]interface{}); ok {
			itemList = append(itemList, item)
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
