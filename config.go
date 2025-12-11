package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 多实例配置
type Config struct {
	Instances []ZabbixInstance `yaml:"instances"`
}

// ZabbixInstance Zabbix实例配置
type ZabbixInstance struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	User     string `yaml:"username,omitempty"`
	Pass     string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
	AuthType string `yaml:"auth_type,omitempty"` // "password" 或 "token"
	Default  bool   `yaml:"default,omitempty"`
}

var AppConfig Config

func LoadConfig() error {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}
