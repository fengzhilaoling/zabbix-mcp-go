package main

import (
	"flag"
	"fmt"
	"zabbix-mcp-go/zabbix"

	"github.com/mark3labs/mcp-go/server"
)

var (
	pool *zabbix.ZabbixPool
)

func main() {
	// 定义命令行参数
	var (
		stdioMode = flag.Bool("stdio", false, "使用stdio传输方式")
		httpMode  = flag.Bool("http", false, "使用HTTP/SSE传输方式")
		port      = flag.Int("port", 5443, "HTTP/SSE监听端口")
	)
	flag.Parse()

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

		// 获取实例信息
		instanceInfo := fmt.Sprintf("实例名称: %s, 地址: %s, 认证方式: %s",
			instance.Name, instance.URL, instance.AuthType)

		// 检测Zabbix版本
		detector := zabbix.NewVersionDetector(client)
		version, err := detector.DetectVersion()
		if err != nil {
			GetSugar().Warnf("%s, 状态: 版本检测失败 - %v", instanceInfo, err)
		} else {
			instanceInfo = fmt.Sprintf("%s, Zabbix版本: %s", instanceInfo, version.String())
		}

		if err := pool.AddInstance(instance.Name, client); err != nil {
			GetSugar().Errorf("%s, 状态: 连接失败 - %v", instanceInfo, err)
			continue
		}

		if instance.Default {
			pool.SetDefault(instance.Name)
		}

		GetSugar().Infof("%s, 状态: 连接成功", instanceInfo)
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

	// 根据参数选择传输方式
	if *stdioMode {
		// 启动stdio服务器
		GetSugar().Info("启动stdio传输方式的MCP服务器...")
		if err := server.ServeStdio(s); err != nil {
			GetSugar().Fatalf("stdio服务器启动失败: %v", err)
		}
	} else if *httpMode {
		// 启动HTTP/SSE服务器
		startHTTPServer(s, *port)
	} else {
		// 默认同时启动两种方式（在不同的goroutine中）
		GetSugar().Info("同时启动stdio和HTTP/SSE传输方式的MCP服务器...")

		// 在后台启动HTTP服务器
		go startHTTPServer(s, *port)

		// 在主线程启动stdio服务器
		if err := server.ServeStdio(s); err != nil {
			GetSugar().Fatalf("stdio服务器启动失败: %v", err)
		}
	}
}

// startHTTPServer 启动HTTP传输服务器（使用SSE）
func startHTTPServer(s *server.MCPServer, port int) {
	addr := fmt.Sprintf(":%d", port)
	GetSugar().Infof("启动HTTP/SSE传输服务器，监听端口: %d", port)
	GetSugar().Infof("MCP端点: http://localhost:%d", port)

	// 使用v0.9.0版本支持的API：创建SSE服务器
	sseServer := server.NewSSEServer(s, fmt.Sprintf("http://localhost:%d", port))
	if err := sseServer.Start(addr); err != nil {
		GetSugar().Fatalf("HTTP/SSE服务器启动失败: %v", err)
	}
}
