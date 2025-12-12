package zabbix

import (
	"fmt"
	"sync"
)

// ZabbixPool Zabbix客户端连接池
type ZabbixPool struct {
	instances       map[string]*ZabbixClient
	defaultInstance string
	mu              sync.RWMutex
}

// NewZabbixPool 创建新的Zabbix连接池
func NewZabbixPool() *ZabbixPool {
	return &ZabbixPool{
		instances: make(map[string]*ZabbixClient),
	}
}

// AddInstance 添加Zabbix实例到池
func (p *ZabbixPool) AddInstance(name string, client *ZabbixClient) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.instances[name]; exists {
		return fmt.Errorf("实例 %s 已存在", name)
	}

	// 尝试登录验证连接
	if err := client.Login(); err != nil {
		return fmt.Errorf("连接实例 %s 失败: %w", name, err)
	}

	p.instances[name] = client

	// 如果这是第一个实例，设为默认
	if p.defaultInstance == "" {
		p.defaultInstance = name
	}

	return nil
}

// RemoveInstance 从池中移除实例
func (p *ZabbixPool) RemoveInstance(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	client, exists := p.instances[name]
	if !exists {
		return fmt.Errorf("实例 %s 不存在", name)
	}

	// 登出并关闭连接
	if err := client.Logout(); err != nil {
		// 记录错误但不阻止移除
		fmt.Printf("警告: 登出实例 %s 失败: %v\n", name, err)
	}

	delete(p.instances, name)

	// 如果移除的是默认实例，重新选择默认实例
	if p.defaultInstance == name {
		p.defaultInstance = ""
		for instanceName := range p.instances {
			p.defaultInstance = instanceName
			break
		}
	}

	return nil
}

// GetClient 获取指定实例的客户端
func (p *ZabbixPool) GetClient(instanceName string) interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if instanceName == "" {
		instanceName = p.defaultInstance
	}

	if client, exists := p.instances[instanceName]; exists {
		return client
	}

	return nil
}

// GetDefaultClient 获取默认实例的客户端
func (p *ZabbixPool) GetDefaultClient() *ZabbixClient {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.defaultInstance != "" {
		return p.instances[p.defaultInstance]
	}

	return nil
}

// SetDefault 设置默认实例
func (p *ZabbixPool) SetDefault(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.instances[name]; !exists {
		return fmt.Errorf("实例 %s 不存在", name)
	}

	p.defaultInstance = name
	return nil
}

// ListInstances 列出所有实例
func (p *ZabbixPool) ListInstances() []map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var instances []map[string]interface{}
	for name, client := range p.instances {
		instance := map[string]interface{}{
			"name":      name,
			"url":       client.URL,
			"user":      client.User,
			"default":   name == p.defaultInstance,
			"connected": client.AuthToken != "",
		}
		instances = append(instances, instance)
	}

	return instances
}

// GetInstanceNames 获取所有实例名称
func (p *ZabbixPool) GetInstanceNames() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	names := make([]string, 0, len(p.instances))
	for name := range p.instances {
		names = append(names, name)
	}

	return names
}

// GetDefaultInstanceName 获取默认实例名称
func (p *ZabbixPool) GetDefaultInstanceName() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.defaultInstance
}

// HealthCheck 检查所有实例的健康状态
func (p *ZabbixPool) HealthCheck() map[string]bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	health := make(map[string]bool)
	for name, client := range p.instances {
		// 尝试调用一个简单的API来检查连接状态
		_, err := client.Call("apiinfo.version", nil)
		health[name] = err == nil
	}

	return health
}

// GetHealthyInstances 获取健康的实例列表
func (p *ZabbixPool) GetHealthyInstances() []string {
	health := p.HealthCheck()
	var healthy []string

	for name, isHealthy := range health {
		if isHealthy {
			healthy = append(healthy, name)
		}
	}

	return healthy
}

// Size 获取池大小
func (p *ZabbixPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.instances)
}

// Clear 清空连接池
func (p *ZabbixPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 登出所有实例
	for name, client := range p.instances {
		if err := client.Logout(); err != nil {
			fmt.Printf("警告: 登出实例 %s 失败: %v\n", name, err)
		}
	}

	// 清空映射
	p.instances = make(map[string]*ZabbixClient)
	p.defaultInstance = ""
}

// InstanceStats 实例统计信息
type InstanceStats struct {
	Name      string
	URL       string
	Connected bool
	AuthToken bool
	IsDefault bool
}

// GetInstanceStats 获取实例统计信息
func (p *ZabbixPool) GetInstanceStats() []InstanceStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var stats []InstanceStats
	for name, client := range p.instances {
		stat := InstanceStats{
			Name:      name,
			URL:       client.URL,
			Connected: client.AuthToken != "",
			AuthToken: client.AuthToken != "",
			IsDefault: name == p.defaultInstance,
		}
		stats = append(stats, stat)
	}

	return stats
}

// MultiInstanceQuery 多实例查询结果
type MultiInstanceQuery struct {
	InstanceName string
	Result       interface{}
	Error        error
}

// QueryAllInstances 查询所有实例
func (p *ZabbixPool) QueryAllInstances(method string, params interface{}) []MultiInstanceQuery {
	p.mu.RLock()
	instances := make(map[string]*ZabbixClient)
	for name, client := range p.instances {
		instances[name] = client
	}
	p.mu.RUnlock()

	var wg sync.WaitGroup
	results := make([]MultiInstanceQuery, 0, len(instances))
	resultChan := make(chan MultiInstanceQuery, len(instances))

	for name, client := range instances {
		wg.Add(1)
		go func(instanceName string, zabbixClient *ZabbixClient) {
			defer wg.Done()

			result, err := zabbixClient.Call(method, params)
			resultChan <- MultiInstanceQuery{
				InstanceName: instanceName,
				Result:       result,
				Error:        err,
			}
		}(name, client)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// QueryHealthyInstances 只查询健康的实例
func (p *ZabbixPool) QueryHealthyInstances(method string, params interface{}) []MultiInstanceQuery {
	healthyInstances := p.GetHealthyInstances()

	p.mu.RLock()
	instances := make(map[string]*ZabbixClient)
	for _, name := range healthyInstances {
		instances[name] = p.instances[name]
	}
	p.mu.RUnlock()

	var wg sync.WaitGroup
	results := make([]MultiInstanceQuery, 0, len(instances))
	resultChan := make(chan MultiInstanceQuery, len(instances))

	for name, client := range instances {
		wg.Add(1)
		go func(instanceName string, zabbixClient *ZabbixClient) {
			defer wg.Done()

			result, err := zabbixClient.Call(method, params)
			resultChan <- MultiInstanceQuery{
				InstanceName: instanceName,
				Result:       result,
				Error:        err,
			}
		}(name, client)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// GetAllInstancesInfo 获取所有实例的详细信息
func (p *ZabbixPool) GetAllInstancesInfo() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]interface{})
	instances := make([]map[string]interface{}, 0)

	for name, client := range p.instances {
		info := client.GetInstanceInfo()
		info["name"] = name
		info["is_default"] = (name == p.defaultInstance)
		instances = append(instances, info)
	}

	result["instances"] = instances
	result["total_count"] = len(instances)
	result["default_instance"] = p.defaultInstance

	return result
}
