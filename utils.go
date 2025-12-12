package main

import "encoding/json"

// 辅助函数
func getStringParam(params map[string]interface{}, key string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return ""
}

func getIntParam(params map[string]interface{}, key string, defaultVal int) int {
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

func getBoolParam(params map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := params[key].(bool); ok {
		return val
	}
	return defaultVal
}

// 辅助：把任意值格式化为缩进 JSON，失败则返回 "null"
func MustJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "null"
	}
	return string(b)
}
