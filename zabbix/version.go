package zabbix

import (
	"fmt"
	"strconv"
	"strings"
)

// VersionInfo Zabbix版本信息
type VersionInfo struct {
	Major      int
	Minor      int
	Patch      int
	Full       string
	APIVersion string
}

// VersionDetector 版本检测器
type VersionDetector struct {
	client *ZabbixClient
}

// NewVersionDetector 创建版本检测器
func NewVersionDetector(client *ZabbixClient) *VersionDetector {
	return &VersionDetector{client: client}
}

// DetectVersion 检测Zabbix版本
func (vd *VersionDetector) DetectVersion() (*VersionInfo, error) {
	// 获取API版本信息
	result, err := vd.client.Call("apiinfo.version", nil)
	if err != nil {
		return nil, fmt.Errorf("获取API版本失败: %w", err)
	}

	apiVersion, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("API版本响应格式错误")
	}

	// 解析版本号
	version, err := vd.parseVersion(apiVersion)
	if err != nil {
		return nil, fmt.Errorf("解析版本号失败: %w", err)
	}

	// 获取完整的版本信息（如果可用）
	if versionInfo, err := vd.getFullVersion(); err == nil {
		version.Full = versionInfo
	}

	version.APIVersion = apiVersion
	return version, nil
}

// parseVersion 解析版本字符串
func (vd *VersionDetector) parseVersion(versionStr string) (*VersionInfo, error) {
	// 移除前缀
	versionStr = strings.TrimPrefix(versionStr, "v")

	// 分割版本号
	parts := strings.Split(versionStr, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("版本格式不正确: %s", versionStr)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("解析主版本号失败: %w", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("解析次版本号失败: %w", err)
	}

	patch := 0
	if len(parts) > 2 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			patch = 0 // 如果解析失败，默认为0
		}
	}

	return &VersionInfo{
		Major: major,
		Minor: minor,
		Patch: patch,
		Full:  versionStr,
	}, nil
}

// getFullVersion 获取完整版本信息
func (vd *VersionDetector) getFullVersion() (string, error) {
	// 尝试获取更详细的版本信息
	params := map[string]interface{}{
		"output": "extend",
	}

	result, err := vd.client.Call("apiinfo.version", params)
	if err != nil {
		return "", err
	}

	if version, ok := result.(string); ok {
		return version, nil
	}

	return "", fmt.Errorf("无法获取完整版本信息")
}

// IsVersionSupported 检查版本是否受支持
func (vd *VersionDetector) IsVersionSupported(minVersion string) (bool, error) {
	currentVersion, err := vd.DetectVersion()
	if err != nil {
		return false, err
	}

	minVer, err := vd.parseVersion(minVersion)
	if err != nil {
		return false, fmt.Errorf("解析最小版本号失败: %w", err)
	}

	return vd.compareVersions(currentVersion, minVer) >= 0, nil
}

// compareVersions 比较两个版本
func (vd *VersionDetector) compareVersions(v1, v2 *VersionInfo) int {
	if v1.Major != v2.Major {
		return v1.Major - v2.Major
	}
	if v1.Minor != v2.Minor {
		return v1.Minor - v2.Minor
	}
	return v1.Patch - v2.Patch
}

// GetCompatibleFeatures 获取当前版本支持的功能
func (vd *VersionDetector) GetCompatibleFeatures() map[string]bool {
	version, err := vd.DetectVersion()
	if err != nil {
		return vd.getDefaultFeatures()
	}

	features := make(map[string]bool)

	// 基础功能（所有版本都支持）
	features["host_management"] = true
	features["item_management"] = true
	features["trigger_management"] = true
	features["template_management"] = true
	features["event_acknowledgment"] = true

	// 版本特定功能
	if version.Major >= 4 {
		features["problem_view"] = true
		features["tag_support"] = true
	}

	if version.Major >= 5 {
		features["advanced_dashboard"] = true
		features["sla_support"] = true
	}

	if version.Major >= 6 {
		features["advanced_analytics"] = true
		features["ai_support"] = true
	}

	// 检查特定的小版本功能
	if version.Major == 4 && version.Minor >= 2 {
		features["improved_api"] = true
	}

	if version.Major == 5 && version.Minor >= 2 {
		features["enhanced_security"] = true
	}

	return features
}

// getDefaultFeatures 获取默认功能集
func (vd *VersionDetector) getDefaultFeatures() map[string]bool {
	return map[string]bool{
		"host_management":      true,
		"item_management":      true,
		"trigger_management":   true,
		"template_management":  true,
		"event_acknowledgment": true,
	}
}

// AdaptAPIParams 根据版本适配API参数
func (vd *VersionDetector) AdaptAPIParams(method string, params map[string]interface{}) map[string]interface{} {
	version, err := vd.DetectVersion()
	if err != nil {
		return params
	}

	adaptedParams := make(map[string]interface{})
	for k, v := range params {
		adaptedParams[k] = v
	}

	// 根据版本调整参数
	switch method {
	case "host.get":
		if version.Major < 4 {
			// 旧版本不支持某些参数
			delete(adaptedParams, "selectTags")
		}
	case "item.get":
		if version.Major < 4 {
			delete(adaptedParams, "selectTags")
			delete(adaptedParams, "selectPreprocessing")
		}
	case "trigger.get":
		if version.Major < 4 {
			delete(adaptedParams, "selectTags")
			delete(adaptedParams, "selectDependencies")
		}
	case "template.get":
		if version.Major < 5 {
			delete(adaptedParams, "selectTags")
		}
	}

	return adaptedParams
}

// GetVersionSpecificEndpoint 获取版本特定的端点
func (vd *VersionDetector) GetVersionSpecificEndpoint(endpoint string) string {
	version, err := vd.DetectVersion()
	if err != nil {
		return endpoint
	}

	// 根据版本返回不同的端点
	switch endpoint {
	case "problem":
		if version.Major >= 4 {
			return "problem.get"
		}
		return "trigger.get" // 旧版本使用trigger.get
	case "sla":
		if version.Major >= 5 {
			return "sla.get"
		}
		return "" // 不支持SLA
	}

	return endpoint
}

// String 返回版本字符串表示
func (v *VersionInfo) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// 在 version.go 中添加更详细的版本特性映射
func (vd *VersionDetector) GetDetailedVersionFeatures() map[string]interface{} {
	version, err := vd.DetectVersion()
	if err != nil {
		// 将 map[string]bool 转换为 map[string]interface{}
		defaultFeatures := vd.getDefaultFeatures()
		result := make(map[string]interface{})
		for k, v := range defaultFeatures {
			result[k] = v
		}
		return result
	}

	features := make(map[string]interface{})

	// API端点支持
	features["endpoints"] = map[string]bool{
		"problem.get":    version.Major >= 4,
		"sla.get":        version.Major >= 5,
		"authentication": version.Major >= 7,
		"connector":      version.Major >= 6,
		"proxygroup":     version.Major >= 7,
	}

	// 参数支持
	features["parameters"] = map[string]bool{
		"selectTags":          version.Major >= 4,
		"selectDependencies":  version.Major >= 4,
		"selectPreprocessing": version.Major >= 4,
		"templateSelectTags":  version.Major >= 5, // template.get
	}

	return features
}

// TestVersionCompatibility 版本兼容性测试函数
func (vd *VersionDetector) TestVersionCompatibility() map[string]interface{} {
	result := make(map[string]interface{})

	// 测试版本检测
	version, err := vd.DetectVersion()
	if err != nil {
		result["version_detection"] = map[string]string{
			"status": "failed",
			"error":  err.Error(),
		}
		return result
	}

	result["version"] = version.String()
	result["features"] = vd.GetCompatibleFeatures()

	// 测试关键API端点
	endpoints := []string{"apiinfo.version", "host.get", "item.get", "trigger.get"}
	if version.Major >= 4 {
		endpoints = append(endpoints, "problem.get")
	}
	if version.Major >= 5 {
		endpoints = append(endpoints, "sla.get")
	}

	endpointTests := make(map[string]string)
	for _, endpoint := range endpoints {
		_, err := vd.client.Call(endpoint, map[string]interface{}{"limit": 1})
		if err != nil {
			endpointTests[endpoint] = "failed: " + err.Error()
		} else {
			endpointTests[endpoint] = "success"
		}
	}

	result["endpoint_tests"] = endpointTests
	return result
}

// 添加版本迁移助手函数
func (vd *VersionDetector) GetMigrationGuide(fromVersion, toVersion string) map[string]interface{} {
	guide := make(map[string]interface{})

	// 解析版本
	fromVer, _ := vd.parseVersion(fromVersion)
	toVer, _ := vd.parseVersion(toVersion)

	if fromVer.Major < 4 && toVer.Major >= 4 {
		guide["problem_management"] = map[string]string{
			"old":         "trigger.get with specific parameters",
			"new":         "problem.get",
			"description": "使用专门的问题管理API替代触发器查询",
		}
	}

	if fromVer.Major < 5 && toVer.Major >= 5 {
		guide["sla_management"] = map[string]string{
			"old":         "service.get with calculations",
			"new":         "sla.get",
			"description": "使用专门的SLA API",
		}
	}

	return guide
}

// TestVersionCompatibility ZabbixClient的版本兼容性测试函数
func (c *ZabbixClient) TestVersionCompatibility() map[string]interface{} {
	detector := NewVersionDetector(c)
	return detector.TestVersionCompatibility()
}
