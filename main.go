package main

import (
	"zabbix-mcp-go/zabbix"

	"github.com/mark3labs/mcp-go/server"
)

var (
	pool *zabbix.ZabbixPool
)

func main() {
	// 初始化日志
	if err := InitLogger(); err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer Sync()

	GetSugar().Info("启动Zabbix MCP服务器")

	// 加载配置
	if err := LoadConfig(); err != nil {
		GetSugar().Fatalf("加载配置失败: %v", err)
	}
	GetSugar().Info("配置加载成功")

	// 初始化Zabbix连接池
	pool = zabbix.NewZabbixPool()
	for _, instance := range AppConfig.Instances {
		client := zabbix.NewZabbixClient(instance.URL, instance.User, instance.Pass)

		// 设置认证方式
		if instance.AuthType == "token" && instance.Token != "" {
			client.SetAuthToken(instance.Token)
		}

		if err := pool.AddInstance(instance.Name, client); err != nil {
			GetSugar().Errorf("添加实例 %s 失败: %v", instance.Name, err)
			continue
		}
		if instance.Default {
			pool.SetDefault(instance.Name)
		}
		GetSugar().Infof("成功添加Zabbix实例: %s", instance.Name)
	}
	GetSugar().Infof("Zabbix连接池初始化完成，共添加 %d 个实例", len(AppConfig.Instances))

	// 创建MCP服务器
	s := server.NewMCPServer(
		"zabbix-mcp-server",
		"1.0.0",
	)
	GetSugar().Info("MCP服务器创建成功")

	// 注册工具
	RegisterTools(s)
	GetSugar().Info("工具注册完成")

	// 启动服务器
	GetSugar().Info("启动MCP服务器...")
	if err := server.ServeStdio(s); err != nil {
		GetSugar().Fatalf("服务器启动失败: %v", err)
	}
}
