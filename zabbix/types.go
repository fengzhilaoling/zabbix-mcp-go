package zabbix

import "fmt"

// JSONRPCRequest JSON-RPC请求结构
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
	Auth    string      `json:"auth,omitempty"`
}

// JSONRPCResponse JSON-RPC响应结构
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *RPCError   `json:"error"`
	ID      int         `json:"id"`
}

// RPCError RPC错误结构
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// Error 实现error接口
func (e *RPCError) Error() string {
	return fmt.Sprintf("Zabbix API Error %d: %s (%s)", e.Code, e.Message, e.Data)
}
