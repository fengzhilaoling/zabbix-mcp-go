package zabbix

import "fmt"

// GetTriggerEvents 获取触发器事件
func (c *ZabbixClient) GetTriggerEvents(triggerID string, limit int) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":              "extend",
		"select_acknowledges": "extend",
		"objectids":           triggerID,
		"source":              0,
		"object":              0,
		"sortfield":           "clock",
		"sortorder":           "DESC",
		"limit":               limit,
	}

	result, err := c.Call("event.get", params)
	if err != nil {
		return nil, err
	}

	events, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var eventList []map[string]interface{}
	for _, e := range events {
		if event, ok := e.(map[string]interface{}); ok {
			eventList = append(eventList, event)
		}
	}

	return eventList, nil
}

// AcknowledgeEvent 确认事件
func (c *ZabbixClient) AcknowledgeEvent(eventID, message string) error {
	params := map[string]interface{}{
		"eventids": eventID,
		"message":  message,
	}

	_, err := c.Call("event.acknowledge", params)
	return err
}

// GetEvents 获取事件列表
func (c *ZabbixClient) GetEvents(params map[string]interface{}) ([]map[string]interface{}, error) {
	if params == nil {
		params = map[string]interface{}{
			"output":    "extend",
			"sortfield": "clock",
			"sortorder": "DESC",
			"limit":     100,
		}
	}

	result, err := c.Call("event.get", params)
	if err != nil {
		return nil, err
	}

	events, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var eventList []map[string]interface{}
	for _, e := range events {
		if event, ok := e.(map[string]interface{}); ok {
			eventList = append(eventList, event)
		}
	}

	return eventList, nil
}

// GetEventByID 根据事件ID获取事件信息
func (c *ZabbixClient) GetEventByID(eventID string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":              "extend",
		"select_acknowledges": "extend",
		"select_alerts":       "extend",
		"eventids":            eventID,
	}

	result, err := c.Call("event.get", params)
	if err != nil {
		return nil, err
	}

	events, ok := result.([]interface{})
	if !ok || len(events) == 0 {
		return nil, fmt.Errorf("事件不存在")
	}

	if event, ok := events[0].(map[string]interface{}); ok {
		return event, nil
	}

	return nil, fmt.Errorf("响应格式错误")
}

// MassAcknowledgeEvents 批量确认事件
func (c *ZabbixClient) MassAcknowledgeEvents(eventIDs []string, message string) error {
	params := map[string]interface{}{
		"eventids": eventIDs,
		"message":  message,
	}

	_, err := c.Call("event.acknowledge", params)
	return err
}

// GetProblemEvents 获取问题事件
func (c *ZabbixClient) GetProblemEvents(params map[string]interface{}) ([]map[string]interface{}, error) {
	if params == nil {
		params = map[string]interface{}{
			"output":    "extend",
			"source":    0,
			"object":    0,
			"sortfield": "clock",
			"sortorder": "DESC",
			"limit":     100,
		}
	}

	result, err := c.Call("problem.get", params)
	if err != nil {
		return nil, err
	}

	problems, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var problemList []map[string]interface{}
	for _, p := range problems {
		if problem, ok := p.(map[string]interface{}); ok {
			problemList = append(problemList, problem)
		}
	}

	return problemList, nil
}
