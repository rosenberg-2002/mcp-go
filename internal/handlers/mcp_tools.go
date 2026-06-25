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

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools đăng ký toàn bộ tool vào MCP Server, chia theo nhóm chức năng.
func RegisterTools(s *server.MCPServer) {
	registerSysInfoTools(s)
	registerFilesystemTools(s)
	registerHTTPTools(s)
	registerDataTools(s)
	registerEnvTools(s)
}

// =============================================================================
// Nhóm 1: System Info — giám sát tài nguyên hệ thống
// Use-case: DevOps dashboard, health check endpoint, capacity planning
// =============================================================================

func registerSysInfoTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("get_system_info",
			mcp.WithDescription("Lấy thông tin hệ thống: OS, kiến trúc CPU, số core, phiên bản Go. Dùng để kiểm tra môi trường deployment."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(sysinfo.GetBasicInfo()), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_memory_stats",
			mcp.WithDescription("Lấy thống kê sử dụng RAM của tiến trình MCP Server. Dùng để phát hiện memory leak trong quá trình vận hành."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(sysinfo.GetMemoryStats()), nil
		},
	)
}

// =============================================================================
// Nhóm 2: File System — thao tác file và thư mục
// Use-case: AI agent đọc/ghi config, phân tích source code, xử lý log file
// =============================================================================

func registerFilesystemTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("list_directory",
			mcp.WithDescription("Liệt kê nội dung (file và thư mục con) tại đường dẫn chỉ định cùng kích thước và thời gian sửa đổi."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Đường dẫn đến thư mục cần liệt kê. Ví dụ: /home/user hoặc C:\\Projects"),
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
			mcp.WithDescription("Đọc và trả về toàn bộ nội dung văn bản của một file (giới hạn 1MB). Dùng để đọc config, log, source code."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Đường dẫn tuyệt đối hoặc tương đối đến file cần đọc."),
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
			mcp.WithDescription("Ghi nội dung vào file (tạo mới nếu chưa tồn tại, ghi đè nếu đã có). Tự động tạo thư mục cha nếu thiếu."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Đường dẫn đến file cần ghi. Ví dụ: /tmp/output.txt"),
			),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("Nội dung cần ghi vào file."),
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
			mcp.WithDescription("Lấy metadata của file hoặc thư mục: kích thước, thời gian sửa đổi, quyền truy cập, đường dẫn tuyệt đối."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Đường dẫn đến file hoặc thư mục cần kiểm tra."),
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
			mcp.WithDescription("Tìm kiếm chuỗi văn bản trong file và trả về số dòng cùng nội dung khớp (không phân biệt hoa/thường). Dùng để tìm lỗi trong log."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Đường dẫn đến file cần tìm kiếm."),
			),
			mcp.WithString("pattern",
				mcp.Required(),
				mcp.Description("Chuỗi hoặc từ khóa cần tìm trong file."),
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
// Nhóm 3: HTTP Client — tích hợp API bên ngoài
// Use-case: Gọi REST API nội bộ, webhook, lấy dữ liệu từ external service
// =============================================================================

func registerHTTPTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("http_get",
			mcp.WithDescription("Thực hiện HTTP GET request và trả về status code + response body. Chỉ hỗ trợ http/https."),
			mcp.WithString("url",
				mcp.Required(),
				mcp.Description("URL cần gọi. Phải bắt đầu bằng http:// hoặc https://"),
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
			mcp.WithDescription("Thực hiện HTTP POST với JSON body và trả về response. Dùng để tạo/cập nhật resource qua REST API. Chỉ hỗ trợ http/https."),
			mcp.WithString("url",
				mcp.Required(),
				mcp.Description("URL endpoint cần gọi."),
			),
			mcp.WithString("body",
				mcp.Required(),
				mcp.Description("JSON string làm request body. Ví dụ: {\"name\": \"Alice\", \"age\": 30}"),
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
// Nhóm 4: Data Utilities — xử lý và biến đổi dữ liệu
// Use-case: ETL pipeline, data normalization, security audit, text processing
// =============================================================================

func registerDataTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("generate_hash",
			mcp.WithDescription("Tạo MD5 và SHA256 hash cho một chuỗi văn bản. Dùng trong kiểm tra tính toàn vẹn dữ liệu hoặc làm cache key."),
			mcp.WithString("text",
				mcp.Required(),
				mcp.Description("Chuỗi cần tạo hash."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(datautils.GenerateHash(req.GetString("text", ""))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("base64_encode",
			mcp.WithDescription("Mã hóa chuỗi văn bản sang Base64. Dùng để truyền dữ liệu nhị phân trong JSON/HTTP header (Basic Auth, JWT)."),
			mcp.WithString("text",
				mcp.Required(),
				mcp.Description("Chuỗi văn bản cần mã hóa."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(datautils.Base64Encode(req.GetString("text", ""))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("base64_decode",
			mcp.WithDescription("Giải mã chuỗi Base64 về văn bản gốc."),
			mcp.WithString("encoded",
				mcp.Required(),
				mcp.Description("Chuỗi Base64 cần giải mã."),
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
			mcp.WithDescription("Định dạng (pretty-print) và validate JSON. Trả về lỗi nếu JSON không hợp lệ. Dùng khi debug API response."),
			mcp.WithString("json",
				mcp.Required(),
				mcp.Description("Chuỗi JSON cần định dạng hoặc validate."),
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
			mcp.WithDescription("Tìm tất cả chuỗi khớp với biểu thức regex trong văn bản. Dùng để trích xuất email, IP, số điện thoại từ log/text."),
			mcp.WithString("pattern",
				mcp.Required(),
				mcp.Description("Biểu thức regex. Ví dụ: \\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b để tìm email."),
			),
			mcp.WithString("text",
				mcp.Required(),
				mcp.Description("Văn bản nguồn cần tìm kiếm."),
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
			mcp.WithDescription("Đếm số ký tự, từ và dòng trong văn bản. Dùng để kiểm tra độ dài nội dung trước khi lưu hoặc gửi đi."),
			mcp.WithString("text",
				mcp.Required(),
				mcp.Description("Đoạn văn bản cần thống kê."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(datautils.WordCount(req.GetString("text", ""))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("text_transform",
			mcp.WithDescription("Chuyển đổi định dạng văn bản. Hỗ trợ: upper, lower, title, reverse, trim, snake_case, camel_case. Dùng để chuẩn hóa tên biến, dữ liệu."),
			mcp.WithString("text",
				mcp.Required(),
				mcp.Description("Văn bản cần chuyển đổi."),
			),
			mcp.WithString("operation",
				mcp.Required(),
				mcp.Description("Phép biến đổi: upper | lower | title | reverse | trim | snake_case | camel_case"),
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
			mcp.WithDescription("Lấy thời gian hiện tại theo timezone chỉ định (RFC3339, Unix timestamp, tuần ISO). Dùng trong ứng dụng đa quốc gia và logging."),
			mcp.WithString("timezone",
				mcp.Required(),
				mcp.Description("Tên timezone theo chuẩn IANA. Ví dụ: UTC, Asia/Ho_Chi_Minh, America/New_York, Europe/London, Asia/Tokyo"),
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
// Nhóm 5: Environment Variables — quản lý cấu hình runtime
// Use-case: Kiểm tra 12-factor app config, debug deployment, CI/CD pipeline
// =============================================================================

func registerEnvTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("get_env_var",
			mcp.WithDescription("Lấy giá trị của một biến môi trường. Các biến nhạy cảm (chứa PASSWORD, TOKEN, KEY...) sẽ được ẩn tự động."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Tên biến môi trường cần lấy. Ví dụ: PATH, GOPATH, HOME, PORT, DB_HOST"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(envtools.GetEnvVar(req.GetString("name", ""))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_env_vars",
			mcp.WithDescription("Liệt kê các biến môi trường theo prefix (để trống để lấy tất cả). Các biến nhạy cảm sẽ bị ẩn giá trị."),
			mcp.WithString("prefix",
				mcp.Description("Prefix để lọc. Ví dụ: 'GO' sẽ lọc GOPATH, GOROOT... Để trống để liệt kê tất cả."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(envtools.ListEnvVars(req.GetString("prefix", ""))), nil
		},
	)

	// Bonus: kiểm tra nhiều biến cùng lúc — pattern hữu ích khi verify deployment checklist
	s.AddTool(
		mcp.NewTool("check_required_env_vars",
			mcp.WithDescription("Kiểm tra xem một danh sách các biến môi trường bắt buộc có tồn tại và không rỗng không. Dùng trong startup check của ứng dụng."),
			mcp.WithString("vars",
				mcp.Required(),
				mcp.Description("Danh sách tên biến cách nhau bởi dấu phẩy. Ví dụ: DB_HOST,DB_PORT,JWT_SECRET"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(checkRequiredEnvVars(req.GetString("vars", ""))), nil
		},
	)
}

// checkRequiredEnvVars kiểm tra danh sách biến môi trường bắt buộc.
func checkRequiredEnvVars(varList string) string {
	names := splitAndTrim(varList, ",")
	if len(names) == 0 {
		return "Danh sách biến rỗng"
	}

	var ok, missing []string
	for _, name := range names {
		if name == "" {
			continue
		}
		val, exists := os.LookupEnv(name)
		if !exists || val == "" {
			missing = append(missing, fmt.Sprintf("  [THIẾU] %s", name))
		} else {
			ok = append(ok, fmt.Sprintf("  [OK]    %s", name))
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Kiểm tra %d biến môi trường bắt buộc:\n", len(names)))
	sb.WriteString(strings.Repeat("─", 50) + "\n")
	for _, s := range ok {
		sb.WriteString(s + "\n")
	}
	for _, s := range missing {
		sb.WriteString(s + "\n")
	}
	sb.WriteString(strings.Repeat("─", 50) + "\n")
	if len(missing) == 0 {
		sb.WriteString(fmt.Sprintf("Kết quả: Tất cả %d biến đều hợp lệ.", len(names)))
	} else {
		sb.WriteString(fmt.Sprintf("Kết quả: %d/%d biến bị thiếu hoặc rỗng.", len(missing), len(names)))
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
