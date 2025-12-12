package zabbix

import "fmt"

// GetTemplates 获取模板列表
func (c *ZabbixClient) GetTemplates() ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":       []string{"templateid", "host", "name", "description"},
		"selectGroups": "extend",
	}

	result, err := c.Call("template.get", params)
	if err != nil {
		return nil, err
	}

	templates, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var templateList []map[string]interface{}
	for _, t := range templates {
		if template, ok := t.(map[string]interface{}); ok {
			templateList = append(templateList, template)
		}
	}

	return templateList, nil
}

// GetTemplateByID 根据模板ID获取模板信息
func (c *ZabbixClient) GetTemplateByID(templateID string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":             "extend",
		"selectGroups":       "extend",
		"selectHosts":        "extend",
		"selectTemplates":    "extend",
		"selectTriggers":     "extend",
		"selectItems":        "extend",
		"selectApplications": "extend",
		"templateids":        templateID,
	}

	result, err := c.Call("template.get", params)
	if err != nil {
		return nil, err
	}

	templates, ok := result.([]interface{})
	if !ok || len(templates) == 0 {
		return nil, fmt.Errorf("模板不存在")
	}

	if template, ok := templates[0].(map[string]interface{}); ok {
		return template, nil
	}

	return nil, fmt.Errorf("响应格式错误")
}

// LinkTemplate 关联模板到主机
func (c *ZabbixClient) LinkTemplate(hostID, templateID string) error {
	params := map[string]interface{}{
		"hosts": []map[string]string{
			{"hostid": hostID},
		},
		"templates": []map[string]string{
			{"templateid": templateID},
		},
	}

	_, err := c.Call("template.massadd", params)
	return err
}

// UnlinkTemplate 从主机移除模板
func (c *ZabbixClient) UnlinkTemplate(hostID, templateID string) error {
	params := map[string]interface{}{
		"hostids":           hostID,
		"templateids_clear": templateID,
	}

	_, err := c.Call("host.massremove", params)
	return err
}

// MassLinkTemplates 批量关联模板到主机
func (c *ZabbixClient) MassLinkTemplates(hostIDs, templateIDs []string) error {
	hosts := make([]map[string]string, len(hostIDs))
	for i, hostID := range hostIDs {
		hosts[i] = map[string]string{"hostid": hostID}
	}

	templates := make([]map[string]string, len(templateIDs))
	for i, templateID := range templateIDs {
		templates[i] = map[string]string{"templateid": templateID}
	}

	params := map[string]interface{}{
		"hosts":     hosts,
		"templates": templates,
	}

	_, err := c.Call("template.massadd", params)
	return err
}

// MassUnlinkTemplates 从主机批量移除模板
func (c *ZabbixClient) MassUnlinkTemplates(hostIDs, templateIDs []string) error {
	params := map[string]interface{}{
		"hostids":           hostIDs,
		"templateids_clear": templateIDs,
	}

	_, err := c.Call("host.massremove", params)
	return err
}

// GetTemplatesByHost 获取主机关联的模板
func (c *ZabbixClient) GetTemplatesByHost(hostID string) ([]map[string]interface{}, error) {
	params := map[string]interface{}{
		"output":  []string{"templateid", "host", "name", "description"},
		"hostids": hostID,
	}

	result, err := c.Call("template.get", params)
	if err != nil {
		return nil, err
	}

	templates, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误")
	}

	var templateList []map[string]interface{}
	for _, t := range templates {
		if template, ok := t.(map[string]interface{}); ok {
			templateList = append(templateList, template)
		}
	}

	return templateList, nil
}

// LinkTemplates 关联模板到主机（单主机版本）
func (c *ZabbixClient) LinkTemplates(hostID string, templateIDs []string) error {
	return c.MassLinkTemplates([]string{hostID}, templateIDs)
}

// UnlinkTemplates 从主机移除模板（单主机版本）
func (c *ZabbixClient) UnlinkTemplates(hostID string, templateIDs []string, clear bool) error {
	if clear {
		return c.MassUnlinkTemplates([]string{hostID}, templateIDs)
	}
	// 如果不清除，使用host.update方法
	params := map[string]interface{}{
		"hostid":          hostID,
		"templates_clear": templateIDs,
	}

	_, err := c.Call("host.update", params)
	return err
}
