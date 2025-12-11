package main

import (
	"log"

	"zabbix-mcp-go/zabbix"

	"github.com/mark3labs/mcp-go/server"
)

var (
	pool *zabbix.ZabbixPool
)

func main() {
	// 加载配置
	if err := LoadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化Zabbix连接池
	pool = zabbix.NewZabbixPool()
	for _, instance := range AppConfig.Instances {
		client := zabbix.NewZabbixClient(instance.URL, instance.User, instance.Pass)

		// 设置认证方式
		if instance.AuthType == "token" && instance.Token != "" {
			client.SetAuthToken(instance.Token)
		}

		if err := pool.AddInstance(instance.Name, client); err != nil {
			log.Printf("添加实例 %s 失败: %v", instance.Name, err)
			continue
		}
		if instance.Default {
			pool.SetDefault(instance.Name)
		}
	}

	// 创建MCP服务器
	s := server.NewMCPServer(
		"zabbix-mcp-server",
		"1.0.0",
	)

	// 注册工具
	RegisterTools(s)

	// 启动服务器
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
