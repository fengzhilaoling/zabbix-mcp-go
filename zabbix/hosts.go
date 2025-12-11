package zabbix

import "fmt"

// GetHosts 获取主机列表
func (c *ZabbixClient) GetHosts(groupID, hostName string) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output": []string{"hostid", "host", "name", "status", "available"},
	}

	if groupID != "" {
		params["groupids"] = groupID
	}

	if hostName != "" {
		params["filter"] = map[string]string{
			"host": hostName,
		}
	}

	result, err := c.Call("host.get", params)
	if err != nil {
		return nil, err
	}

	hosts, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var hostList []map[string]interface{}
	for _, h := range hosts {
		if host, ok := h.(map[string]interface{}); ok {
			hostList = append(hostList, host)
		}
	}

	return hostList, nil
}

// GetHostByName 根据主机名获取主机信息
func (c *ZabbixClient) GetHostByName(hostName string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"output": "extend",
		"filter": map[string]string{
			"host": hostName,
		},
		"selectGroups":     "extend",
		"selectInterfaces": "extend",
	}

	result, err := c.Call("host.get", params)
	if err != nil {
		return nil, err
	}

	hosts, ok := result.([]interface{})
	if !ok || len(hosts) == 0 {
		return nil, fmt.Errorf("主机不存在")
	}

	if host, ok := hosts[0].(map[string]interface{}); ok {
		return host, nil
	}

	return nil, fmt.Errorf("响应格式错误")
}

// CreateHost 创建主机
func (c *ZabbixClient) CreateHost(hostName, groupID, interfaceIP string) (string, error) {
	params := map[string]interface{}{
		"host": hostName,
		"interfaces": []map[string]interface{}{
			{
				"type":  1,
				"main":  1,
				"useip": 1,
				"ip":    interfaceIP,
				"dns":   "",
				"port":  "10050",
			},
		},
		"groups": []map[string]interface{}{
			{
				"groupid": groupID,
			},
		},
	}

	result, err := c.Call("host.create", params)
	if err != nil {
		return "", err
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误")
	}

	hostIDs, ok := response["hostids"].([]interface{})
	if !ok || len(hostIDs) == 0 {
		return "", fmt.Errorf("创建主机失败")
	}

	return hostIDs[0].(string), nil
}

// DeleteHost 删除主机
func (c *ZabbixClient) DeleteHost(hostID string) error {
	params := []string{hostID}
	_, err := c.Call("host.delete", params)
	return err
}

// UpdateHost 更新主机信息
func (c *ZabbixClient) UpdateHost(hostID string, params map[string]interface{}) error {
	params["hostid"] = hostID
	_, err := c.Call("host.update", params)
	return err
}

// GetHostGroups 获取主机组列表
func (c *ZabbixClient) GetHostGroups() ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output": []string{"groupid", "name", "internal"},
	}

	result, err := c.Call("hostgroup.get", params)
	if err != nil {
		return nil, err
	}

	groups, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var groupList []map[string]interface{}
	for _, g := range groups {
		if group, ok := g.(map[string]interface{}); ok {
			groupList = append(groupList, group)
		}
	}

	return groupList, nil
}
