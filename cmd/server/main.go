package main

import (
	"fmt"

	"github.com/rosenberg-2002/mcp-go/internal/handlers"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Khởi tạo server
	s := server.NewMCPServer("SysInfoManager", "1.0.0")

	// Ủy quyền việc đăng ký Tools cho package handlers
	handlers.RegisterTools(s)

	// Lắng nghe I/O
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Lỗi khởi chạy server: %v\n", err)
	}
}