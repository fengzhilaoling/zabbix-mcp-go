package main

import (
	"zabbix-mcp-go/handler"

	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
	// 使用handler包中的注册函数
	handler.RegisterTools(s)
}
