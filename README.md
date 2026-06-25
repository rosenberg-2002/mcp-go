# 🛠️ MCP-Go — Multi-Tool MCP Server

A feature-rich [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server built in Go, providing **23 utility tools** for AI assistants like Claude, Gemini, and others.

## ✨ Features

| Category | Tools | Description |
|---|---|---|
| 🖥️ **System Info** | `get_system_info`, `get_memory_stats` | OS details, CPU, RAM monitoring |
| 📁 **File System** | `list_directory`, `read_file_content`, `write_file_content`, `get_file_info`, `search_in_file` | Read, write, search files & directories |
| 🌐 **HTTP Client** | `http_get`, `http_post_json` | Make HTTP requests to external APIs |
| 🔧 **Data Utils** | `generate_hash`, `base64_encode`, `base64_decode`, `format_json`, `regex_match`, `word_count`, `text_transform`, `get_current_time` | Encoding, hashing, text processing |
| 🔑 **Env Tools** | `get_env_var`, `list_env_vars`, `check_required_env_vars` | Environment variable management (auto-hides secrets) |
| 🎬 **YouTube** | `get_youtube_info`, `get_youtube_transcript`, `list_youtube_languages` | Video metadata, transcripts & language listing |

---

## 🚀 Installation

### Option 1: Install with Go (Recommended)

If you have Go installed:

```bash
go install github.com/rosenberg-2002/mcp-go/cmd/server@latest
```

The binary will be placed in your `$GOPATH/bin` (or `$HOME/go/bin` by default).

### Option 2: Download Pre-built Binary

Go to the [Releases](https://github.com/rosenberg-2002/mcp-go/releases) page and download the binary for your OS:

| OS | Architecture | File |
|---|---|---|
| Windows | x64 | `mcp-server-windows-amd64.exe` |
| macOS | Intel | `mcp-server-darwin-amd64` |
| macOS | Apple Silicon | `mcp-server-darwin-arm64` |
| Linux | x64 | `mcp-server-linux-amd64` |

### Option 3: Build from Source

```bash
git clone https://github.com/rosenberg-2002/mcp-go.git
cd mcp-go
go build -o mcp-server ./cmd/server
```

---

## ⚙️ Configuration

### Claude Desktop

Add to your Claude Desktop config file:

**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "mcp-go": {
      "command": "C:\\path\\to\\mcp-server.exe"
    }
  }
}
```

> 💡 If installed via `go install`, use the full path to the binary in your `$GOPATH/bin`.

### Cursor / Other MCP Clients

Add this to your MCP client configuration:

```json
{
  "mcpServers": {
    "mcp-go": {
      "command": "/path/to/mcp-server"
    }
  }
}
```

---

## 🏗️ Building for All Platforms

To cross-compile binaries for all platforms:

```bash
# Windows (x64)
GOOS=windows GOARCH=amd64 go build -o dist/mcp-server-windows-amd64.exe ./cmd/server

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o dist/mcp-server-darwin-amd64 ./cmd/server

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o dist/mcp-server-darwin-arm64 ./cmd/server

# Linux (x64)
GOOS=linux GOARCH=amd64 go build -o dist/mcp-server-linux-amd64 ./cmd/server
```

Or use the included build script:

```bash
# PowerShell (Windows)
.\build.ps1

# Bash (macOS/Linux)
./build.sh
```

---

## 📁 Project Structure

```
mcp-go/
├── cmd/server/          # Entry point (main.go)
├── internal/
│   ├── handlers/        # MCP tool registration & request routing
│   ├── datautils/       # Hashing, encoding, JSON, regex, time utils
│   ├── envtools/        # Environment variable tools
│   ├── filesystem/      # File I/O operations
│   ├── httpclient/      # HTTP GET/POST client
│   ├── sysinfo/         # System & memory info
│   └── youtube/         # YouTube video info & transcripts
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## 📜 License

[MIT](LICENSE) — use it however you like.
