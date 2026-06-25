# GitHub Repository Professional Updates

Here are the professional updates for your GitHub repository. Since I don't have access to your Personal Access Token (PAT) to update these directly via the API, you can copy and paste these into your repository settings, or let me know if you'd prefer a script to automate this!

## 1. Project Description (About Section)

**Current (Vietnamese):**
> Xây dựng một Model Context Protocol (MCP) Server cung cấp một tập hợp các công cụ (tools) đa năng cho các AI Agent (như Claude, Gemini) gọi và tương tác với hệ thống cục bộ.

**Professional Update (English):**
> A versatile Model Context Protocol (MCP) server built in Go, providing a comprehensive suite of local system tools for AI agents like Claude and Gemini.

---

## 2. Release Notes

**Target Release:** `Releases` (MCP Server exe file)

**Suggested Release Title:** 
> MCP-Go Server (Initial Release)

**Suggested Release Description / Body:**

```markdown
We are excited to announce the release of the **MCP-Go Server**, providing a robust suite of local system tools for AI assistants. This release includes pre-compiled binaries for Windows, macOS, and Linux to ensure seamless integration with your favorite MCP clients like Claude Desktop and Cursor.

### 🚀 Getting Started

1. **Download** the appropriate binary for your operating system from the assets below.
2. **Configure** your MCP client by adding the server path to your configuration file.

**Example for Claude Desktop:**
`​`​`json
{
  "mcpServers": {
    "mcp-go": {
      "command": "/absolute/path/to/downloaded/mcp-server-executable"
    }
  }
}
`​`​`

For detailed installation instructions, cross-compiling, and tool documentation, please refer to our [README](https://github.com/rosenberg-2002/mcp-go/blob/main/README.md).
```
