package handler

import (
	"go.uber.org/zap"
)

// 全局变量和函数声明，这些需要在主程序中初始化
var (
	pool     ClientPool
	sugar    *zap.SugaredLogger
	mustJSON func(v interface{}) string

	// 类型转换函数
	getZabbixClient func(client interface{}) ZabbixClient
)

// ClientPool 接口定义
type ClientPool interface {
	GetClient(instanceName string) interface{}
	ListInstances() []map[string]interface{}
	SetDefault(instanceName string) error
}

// ZabbixClient 接口定义
type ZabbixClient interface {
	// 主机相关
	GetHostsWithPagination(groupID, hostName string, page, pageSize int) ([]map[string]interface{}, int, error)
	GetHosts(groupID, hostName string) ([]map[string]interface{}, error)
	GetHostByName(hostName string) (map[string]interface{}, error)
	GetHostByNameLite(hostName string) (map[string]interface{}, error)
	CreateHost(hostName, groupID, interfaceIP string) (string, error)
	DeleteHost(hostID string) error

	// 监控项相关
	GetItems(hostID, itemNameFilter string) ([]map[string]interface{}, error)
	GetItemInfo(itemID string) (map[string]interface{}, error)
	GetItemData(itemID string, history, limit int) ([]map[string]interface{}, error)
	GetItemDataWithTimeRange(itemID string, history int, timeFrom, timeTill string) ([]map[string]interface{}, error)
	CreateItem(hostID, itemName, key, itemType, valueType, delay string) (string, error)

	// 触发器相关
	GetTriggers(hostID string, active bool) ([]map[string]interface{}, error)
	GetTriggerEvents(triggerID string, limit int) ([]map[string]interface{}, error)
	MassAcknowledgeEvents(eventIDs []string, message string) error

	// 模板相关
	GetTemplates() ([]map[string]interface{}, error)
	GetTemplatesByHost(hostID string) ([]map[string]interface{}, error)
	LinkTemplates(hostID string, templateIDs []string) error
	UnlinkTemplates(hostID string, templateIDs []string, clear bool) error
}

// SetDependencies 设置依赖项，由主程序调用
func SetDependencies(p ClientPool, logger *zap.SugaredLogger, jsonFunc func(interface{}) string, clientConverter func(interface{}) ZabbixClient) {
	pool = p
	sugar = logger
	mustJSON = jsonFunc
	getZabbixClient = clientConverter
}

// GetSugar 获取日志记录器
func GetSugar() *zap.SugaredLogger {
	return sugar
}

// MustJSON 将对象转换为JSON字符串
func MustJSON(v interface{}) string {
	if mustJSON != nil {
		return mustJSON(v)
	}
	return ""
}
