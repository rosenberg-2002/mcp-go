package main

import (
	"fmt"

	"github.com/rosenberg-2002/mcp-go/internal/handlers"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Initialize the MCP server
	s := server.NewMCPServer("mcp-go", "1.0.0")

	// Delegate tool registration to the handlers package
	handlers.RegisterTools(s)

	// Start listening on stdio
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server startup error: %v\n", err)
	}
}
