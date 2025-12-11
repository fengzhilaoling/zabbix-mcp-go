package zabbix

import "fmt"

// GetTriggers 获取触发器
func (c *ZabbixClient) GetTriggers(hostID string, active bool) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":      "extend",
		"selectHosts": "extend",
		"selectItems": "extend",
	}

	if hostID != "" {
		params["hostids"] = hostID
	}

	if active {
		params["filter"] = map[string]interface{}{
			"status": 0,
		}
	}

	result, err := c.Call("trigger.get", params)
	if err != nil {
		return nil, err
	}

	triggers, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var triggerList []map[string]interface{}
	for _, t := range triggers {
		if trigger, ok := t.(map[string]interface{}); ok {
			triggerList = append(triggerList, trigger)
		}
	}

	return triggerList, nil
}

// GetTriggerByID 根据触发器ID获取触发器信息
func (c *ZabbixClient) GetTriggerByID(triggerID string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":             "extend",
		"selectHosts":        "extend",
		"selectItems":        "extend",
		"selectDependencies": "extend",
		"triggerids":         triggerID,
	}

	result, err := c.Call("trigger.get", params)
	if err != nil {
		return nil, err
	}

	triggers, ok := result.([]interface{})
	if !ok || len(triggers) == 0 {
		return nil, fmt.Errorf("触发器不存在")
	}

	if trigger, ok := triggers[0].(map[string]interface{}); ok {
		return trigger, nil
	}

	return nil, fmt.Errorf("响应格式错误")
}

// CreateTrigger 创建触发器
func (c *ZabbixClient) CreateTrigger(description, expression string, priority int) (string, error) {
	params := map[string]interface{}{
		"description": description,
		"expression":  expression,
		"priority":    priority,
		"status":      0, // 启用状态
	}

	result, err := c.Call("trigger.create", params)
	if err != nil {
		return "", err
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误")
	}

	triggerIDs, ok := response["triggerids"].([]interface{})
	if !ok || len(triggerIDs) == 0 {
		return "", fmt.Errorf("创建触发器失败")
	}

	return triggerIDs[0].(string), nil
}

// UpdateTrigger 更新触发器
func (c *ZabbixClient) UpdateTrigger(triggerID string, params map[string]interface{}) error {
	params["triggerid"] = triggerID
	_, err := c.Call("trigger.update", params)
	return err
}

// DeleteTrigger 删除触发器
func (c *ZabbixClient) DeleteTrigger(triggerID string) error {
	params := []string{triggerID}
	_, err := c.Call("trigger.delete", params)
	return err
}

// EnableTrigger 启用触发器
func (c *ZabbixClient) EnableTrigger(triggerID string) error {
	params := map[string]interface{}{
		"triggerid": triggerID,
		"status":    0,
	}
	_, err := c.Call("trigger.update", params)
	return err
}

// DisableTrigger 禁用触发器
func (c *ZabbixClient) DisableTrigger(triggerID string) error {
	params := map[string]interface{}{
		"triggerid": triggerID,
		"status":    1,
	}
	_, err := c.Call("trigger.update", params)
	return err
}
