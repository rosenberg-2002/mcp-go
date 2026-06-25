package handlers

import (
"context"
"fmt"
"os"
"strings"

"github.com/rosenberg-2002/mcp-go/internal/datautils"
"github.com/rosenberg-2002/mcp-go/internal/envtools"
"github.com/rosenberg-2002/mcp-go/internal/filesystem"
"github.com/rosenberg-2002/mcp-go/internal/httpclient"
"github.com/rosenberg-2002/mcp-go/internal/sysinfo"
"github.com/rosenberg-2002/mcp-go/internal/youtube"

"github.com/mark3labs/mcp-go/mcp"
"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all tools into the MCP server, grouped by category.
func RegisterTools(s *server.MCPServer) {
registerSysInfoTools(s)
registerFilesystemTools(s)
registerHTTPTools(s)
registerDataTools(s)
registerEnvTools(s)
registerYouTubeTools(s)
}

// =============================================================================
// Group 1: System Info — monitor system resources
// Use-case: DevOps dashboards, health check endpoints, capacity planning
// =============================================================================

func registerSysInfoTools(s *server.MCPServer) {
s.AddTool(
mcp.NewTool("get_system_info",
mcp.WithDescription("Returns system information: OS, CPU architecture, core count, and Go version. Useful for verifying the deployment environment."),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
return mcp.NewToolResultText(sysinfo.GetBasicInfo()), nil
},
)

s.AddTool(
mcp.NewTool("get_memory_stats",
mcp.WithDescription("Returns RAM usage statistics for the MCP server process. Useful for detecting memory leaks during operation."),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
return mcp.NewToolResultText(sysinfo.GetMemoryStats()), nil
},
)
}

// =============================================================================
// Group 2: File System — file and directory operations
// Use-case: AI agents reading/writing configs, analyzing source code, processing log files
// =============================================================================

func registerFilesystemTools(s *server.MCPServer) {
s.AddTool(
mcp.NewTool("list_directory",
mcp.WithDescription("Lists the contents (files and subdirectories) at the specified path, including size and last-modified time."),
mcp.WithString("path",
mcp.Required(),
mcp.Description("Path to the directory to list. Example: /home/user or C:\\Projects"),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
path := req.GetString("path", "")
result, err := filesystem.ListDirectory(path)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("read_file_content",
mcp.WithDescription("Reads and returns the full text content of a file (limit: 1 MB). Useful for reading configs, logs, and source code."),
mcp.WithString("path",
mcp.Required(),
mcp.Description("Absolute or relative path to the file to read."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
path := req.GetString("path", "")
result, err := filesystem.ReadFileContent(path)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("write_file_content",
mcp.WithDescription("Writes content to a file (creates if not exists, overwrites if it does). Automatically creates missing parent directories."),
mcp.WithString("path",
mcp.Required(),
mcp.Description("Path to the file to write. Example: /tmp/output.txt"),
),
mcp.WithString("content",
mcp.Required(),
mcp.Description("Content to write to the file."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
path := req.GetString("path", "")
content := req.GetString("content", "")
result, err := filesystem.WriteFileContent(path, content)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("get_file_info",
mcp.WithDescription("Returns metadata for a file or directory: size, last-modified time, permissions, and absolute path."),
mcp.WithString("path",
mcp.Required(),
mcp.Description("Path to the file or directory to inspect."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
path := req.GetString("path", "")
result, err := filesystem.GetFileInfo(path)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("search_in_file",
mcp.WithDescription("Searches for a text string in a file and returns matching line numbers and content (case-insensitive). Useful for finding errors in log files."),
mcp.WithString("path",
mcp.Required(),
mcp.Description("Path to the file to search."),
),
mcp.WithString("pattern",
mcp.Required(),
mcp.Description("String or keyword to search for in the file."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
path := req.GetString("path", "")
pattern := req.GetString("pattern", "")
result, err := filesystem.SearchInFile(path, pattern)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)
}

// =============================================================================
// Group 3: HTTP Client — integrate with external APIs
// Use-case: Calling internal REST APIs, webhooks, fetching data from external services
// =============================================================================

func registerHTTPTools(s *server.MCPServer) {
s.AddTool(
mcp.NewTool("http_get",
mcp.WithDescription("Performs an HTTP GET request and returns the status code and response body. Only http/https is supported."),
mcp.WithString("url",
mcp.Required(),
mcp.Description("URL to request. Must start with http:// or https://"),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
rawURL := req.GetString("url", "")
result, err := httpclient.Get(rawURL)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("http_post_json",
mcp.WithDescription("Performs an HTTP POST with a JSON body and returns the response. Use to create or update resources via REST API. Only http/https is supported."),
mcp.WithString("url",
mcp.Required(),
mcp.Description("Endpoint URL to call."),
),
mcp.WithString("body",
mcp.Required(),
mcp.Description(`JSON string as the request body. Example: {"name": "Alice", "age": 30}`),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
rawURL := req.GetString("url", "")
body := req.GetString("body", "")
result, err := httpclient.PostJSON(rawURL, body)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)
}

// =============================================================================
// Group 4: Data Utilities — process and transform data
// Use-case: ETL pipelines, data normalization, security auditing, text processing
// =============================================================================

func registerDataTools(s *server.MCPServer) {
s.AddTool(
mcp.NewTool("generate_hash",
mcp.WithDescription("Generates MD5 and SHA256 hashes for a given string. Useful for data integrity checks or cache keys."),
mcp.WithString("text",
mcp.Required(),
mcp.Description("String to hash."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
return mcp.NewToolResultText(datautils.GenerateHash(req.GetString("text", ""))), nil
},
)

s.AddTool(
mcp.NewTool("base64_encode",
mcp.WithDescription("Encodes a string to Base64. Useful for transmitting binary data in JSON/HTTP headers (Basic Auth, JWT)."),
mcp.WithString("text",
mcp.Required(),
mcp.Description("String to encode."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
return mcp.NewToolResultText(datautils.Base64Encode(req.GetString("text", ""))), nil
},
)

s.AddTool(
mcp.NewTool("base64_decode",
mcp.WithDescription("Decodes a Base64 string back to plain text."),
mcp.WithString("encoded",
mcp.Required(),
mcp.Description("Base64 string to decode."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
result, err := datautils.Base64Decode(req.GetString("encoded", ""))
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("format_json",
mcp.WithDescription("Pretty-prints and validates a JSON string. Returns an error if the JSON is invalid. Useful for debugging API responses."),
mcp.WithString("json",
mcp.Required(),
mcp.Description("JSON string to format or validate."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
result, err := datautils.FormatJSON(req.GetString("json", ""))
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("regex_match",
mcp.WithDescription(`Finds all strings matching a regex pattern in the given text. Useful for extracting emails, IPs, or phone numbers from logs.`),
mcp.WithString("pattern",
mcp.Required(),
mcp.Description(`Regex pattern. Example: \b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b to find emails.`),
),
mcp.WithString("text",
mcp.Required(),
mcp.Description("Source text to search."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
result, err := datautils.RegexMatch(req.GetString("pattern", ""), req.GetString("text", ""))
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("word_count",
mcp.WithDescription("Counts characters, words, and lines in a text. Useful for checking content length before saving or sending."),
mcp.WithString("text",
mcp.Required(),
mcp.Description("Text to analyze."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
return mcp.NewToolResultText(datautils.WordCount(req.GetString("text", ""))), nil
},
)

s.AddTool(
mcp.NewTool("text_transform",
mcp.WithDescription("Transforms text format. Supported operations: upper, lower, title, reverse, trim, snake_case, camel_case. Useful for normalizing variable names and data."),
mcp.WithString("text",
mcp.Required(),
mcp.Description("Text to transform."),
),
mcp.WithString("operation",
mcp.Required(),
mcp.Description("Transformation: upper | lower | title | reverse | trim | snake_case | camel_case"),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
result, err := datautils.TextTransform(req.GetString("text", ""), req.GetString("operation", ""))
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("get_current_time",
mcp.WithDescription("Returns the current time in the specified timezone (RFC3339, Unix timestamp, ISO week). Useful for multi-region apps and standardized logging."),
mcp.WithString("timezone",
mcp.Required(),
mcp.Description("IANA timezone name. Example: UTC, America/New_York, Europe/London, Asia/Tokyo"),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
result, err := datautils.GetCurrentTime(req.GetString("timezone", ""))
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)
}

// =============================================================================
// Group 5: Environment Variables — manage runtime configuration
// Use-case: Verify 12-factor app config, debug deployments, CI/CD pipelines
// =============================================================================

func registerEnvTools(s *server.MCPServer) {
s.AddTool(
mcp.NewTool("get_env_var",
mcp.WithDescription("Returns the value of an environment variable. Sensitive variables (containing PASSWORD, TOKEN, KEY, etc.) are automatically masked."),
mcp.WithString("name",
mcp.Required(),
mcp.Description("Name of the environment variable. Example: PATH, GOPATH, HOME, PORT, DB_HOST"),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
return mcp.NewToolResultText(envtools.GetEnvVar(req.GetString("name", ""))), nil
},
)

s.AddTool(
mcp.NewTool("list_env_vars",
mcp.WithDescription("Lists environment variables filtered by prefix (leave empty to list all). Sensitive variable values are masked."),
mcp.WithString("prefix",
mcp.Description("Prefix to filter by. Example: 'GO' filters GOPATH, GOROOT, etc. Leave empty to list all."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
return mcp.NewToolResultText(envtools.ListEnvVars(req.GetString("prefix", ""))), nil
},
)

// Check multiple variables at once — useful for verifying deployment checklists
s.AddTool(
mcp.NewTool("check_required_env_vars",
mcp.WithDescription("Checks whether a list of required environment variables exist and are non-empty. Useful for application startup checks."),
mcp.WithString("vars",
mcp.Required(),
mcp.Description("Comma-separated list of variable names. Example: DB_HOST,DB_PORT,JWT_SECRET"),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
return mcp.NewToolResultText(checkRequiredEnvVars(req.GetString("vars", ""))), nil
},
)
}

// checkRequiredEnvVars checks a list of required environment variables.
func checkRequiredEnvVars(varList string) string {
names := splitAndTrim(varList, ",")
if len(names) == 0 {
return "Variable list is empty"
}

var ok, missing []string
for _, name := range names {
if name == "" {
continue
}
val, exists := os.LookupEnv(name)
if !exists || val == "" {
missing = append(missing, fmt.Sprintf("  [MISSING] %s", name))
} else {
ok = append(ok, fmt.Sprintf("  [OK]      %s", name))
}
}

var sb strings.Builder
sb.WriteString(fmt.Sprintf("Checking %d required environment variables:\n", len(names)))
sb.WriteString(strings.Repeat("─", 50) + "\n")
for _, s := range ok {
sb.WriteString(s + "\n")
}
for _, s := range missing {
sb.WriteString(s + "\n")
}
sb.WriteString(strings.Repeat("─", 50) + "\n")
if len(missing) == 0 {
sb.WriteString(fmt.Sprintf("Result: All %d variables are set.", len(names)))
} else {
sb.WriteString(fmt.Sprintf("Result: %d/%d variables are missing or empty.", len(missing), len(names)))
}
return sb.String()
}

func splitAndTrim(s, sep string) []string {
parts := strings.Split(s, sep)
result := make([]string, 0, len(parts))
for _, p := range parts {
if t := strings.TrimSpace(p); t != "" {
result = append(result, t)
}
}
return result
}

// =============================================================================
// Group 6: YouTube — retrieve video information from YouTube links
// Use-case: Extracting video metadata, content analysis, information aggregation
// =============================================================================

func registerYouTubeTools(s *server.MCPServer) {
s.AddTool(
mcp.NewTool("get_youtube_info",
mcp.WithDescription("Retrieves detailed information for a YouTube video from its URL: title, channel, description, view count, duration, publish date, and thumbnail. Supports multiple URL formats: youtube.com/watch?v=, youtu.be/, youtube.com/shorts/, youtube.com/embed/."),
mcp.WithString("url",
mcp.Required(),
mcp.Description("YouTube video URL. Example: https://www.youtube.com/watch?v=dQw4w9WgXcQ or https://youtu.be/dQw4w9WgXcQ"),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
rawURL := req.GetString("url", "")
result, err := youtube.GetVideoInfo(rawURL)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("get_youtube_transcript",
mcp.WithDescription("Retrieves the transcript (subtitles) of a YouTube video. Returns full text with timestamps. Supports language selection. Useful for reading video content without watching."),
mcp.WithString("url",
mcp.Required(),
mcp.Description("YouTube video URL."),
),
mcp.WithString("lang",
mcp.Description("Subtitle language code (e.g. en, ja, ko). Leave empty to auto-select the best available track."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
rawURL := req.GetString("url", "")
lang := req.GetString("lang", "")
result, err := youtube.GetTranscript(rawURL, lang)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)

s.AddTool(
mcp.NewTool("list_youtube_languages",
mcp.WithDescription("Lists available subtitle/transcript languages for a YouTube video. Call this before get_youtube_transcript to see which languages are available."),
mcp.WithString("url",
mcp.Required(),
mcp.Description("YouTube video URL."),
),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
rawURL := req.GetString("url", "")
result, err := youtube.GetAvailableLanguages(rawURL)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(result), nil
},
)
}
