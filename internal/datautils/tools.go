package datautils

import (
	"crypto/md5" //nolint:gosec // dùng cho mục đích giáo dục, không dùng cho bảo mật
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// GenerateHash tạo MD5 và SHA256 hash từ chuỗi đầu vào.
// Ứng dụng thực tế: Kiểm tra tính toàn vẹn file, cache key, content-addressable storage.
func GenerateHash(text string) string {
	md5Sum := md5.Sum([]byte(text)) //nolint:gosec
	sha256Sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf(
		"Input  : %q\n%s\nMD5    : %x\nSHA256 : %x\n\nLưu ý: MD5 không nên dùng cho mục đích bảo mật (dùng SHA256 thay thế).",
		text, strings.Repeat("─", 50), md5Sum, sha256Sum,
	)
}

// Base64Encode mã hóa văn bản sang Base64.
// Ứng dụng thực tế: Truyền dữ liệu nhị phân qua HTTP header/JSON, Basic Auth header.
func Base64Encode(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// Base64Decode giải mã chuỗi Base64 về văn bản gốc.
func Base64Decode(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("chuỗi Base64 không hợp lệ: %w", err)
	}
	return string(data), nil
}

// FormatJSON định dạng (pretty-print) và validate JSON.
// Ứng dụng thực tế: Debug API response, hiển thị config file dễ đọc hơn.
func FormatJSON(jsonStr string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("JSON không hợp lệ: %w", err)
	}
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("không thể format JSON: %w", err)
	}
	return string(formatted), nil
}

// RegexMatch tìm tất cả chuỗi khớp với biểu thức regex trong văn bản.
// Ứng dụng thực tế: Trích xuất email/số điện thoại/IP từ log, validate định dạng dữ liệu.
func RegexMatch(pattern, text string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("biểu thức regex không hợp lệ: %w", err)
	}
	matches := re.FindAllString(text, -1)
	if len(matches) == 0 {
		return fmt.Sprintf("Không tìm thấy kết quả nào khớp với pattern: %q", pattern), nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Pattern : %s\nTìm thấy %d kết quả:\n%s\n", pattern, len(matches), strings.Repeat("─", 40)))
	for i, m := range matches {
		sb.WriteString(fmt.Sprintf("  [%d] %q\n", i+1, m))
	}
	return sb.String(), nil
}

// WordCount đếm ký tự, từ và dòng trong văn bản.
// Ứng dụng thực tế: Kiểm tra độ dài nội dung, thống kê tài liệu.
func WordCount(text string) string {
	lines := strings.Count(text, "\n") + 1
	words := len(strings.Fields(text))
	chars := len([]rune(text))
	charsNoSpace := len([]rune(strings.ReplaceAll(text, " ", "")))
	return fmt.Sprintf(
		"Thống kê văn bản:\n%s\n  Ký tự (có khoảng trắng) : %d\n  Ký tự (không khoảng trắng): %d\n  Từ                       : %d\n  Dòng                     : %d",
		strings.Repeat("─", 40), chars, charsNoSpace, words, lines,
	)
}

// TextTransform chuyển đổi định dạng văn bản.
// Ứng dụng thực tế: Chuẩn hóa tên biến, chuẩn bị dữ liệu trước khi lưu DB.
func TextTransform(text, operation string) (string, error) {
	switch strings.ToLower(operation) {
	case "upper":
		return strings.ToUpper(text), nil
	case "lower":
		return strings.ToLower(text), nil
	case "title":
		return toTitleCase(text), nil
	case "reverse":
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	case "trim":
		return strings.TrimSpace(text), nil
	case "snake_case":
		return toSnakeCase(text), nil
	case "camel_case":
		return toCamelCase(text), nil
	default:
		return "", fmt.Errorf(
			"operation không hợp lệ: '%s'\nCác operation được hỗ trợ: upper, lower, title, reverse, trim, snake_case, camel_case",
			operation,
		)
	}
}

// GetCurrentTime trả về thời gian hiện tại theo timezone chỉ định.
// Ứng dụng thực tế: Ứng dụng đa quốc gia, logging timestamp chuẩn, báo cáo theo vùng.
func GetCurrentTime(timezone string) (string, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf(
			"timezone không hợp lệ: '%s'\nVí dụ hợp lệ: UTC, Asia/Ho_Chi_Minh, America/New_York, Europe/London, Asia/Tokyo",
			timezone,
		)
	}
	now := time.Now().In(loc)
	year, week := now.ISOWeek()
	_, month, day := now.Date()

	return fmt.Sprintf(
		"Thời gian hiện tại (%s)\n%s\n  RFC3339         : %s\n  Định dạng VN    : %02d/%02d/%d %02d:%02d:%02d\n  Unix Timestamp  : %d\n  Tuần ISO        : Tuần %d năm %d\n  Ngày trong năm  : %d/%d",
		timezone,
		strings.Repeat("─", 50),
		now.Format(time.RFC3339),
		day, int(month), now.Year(),
		now.Hour(), now.Minute(), now.Second(),
		now.Unix(),
		week, year,
		now.YearDay(), daysInYear(now.Year()),
	), nil
}

// --- Helpers nội bộ ---

func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		runes := []rune(w)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

func toSnakeCase(s string) string {
	var sb strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			sb.WriteRune('_')
		}
		sb.WriteRune(unicode.ToLower(r))
	}
	return strings.ReplaceAll(sb.String(), " ", "_")
}

func toCamelCase(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	result := strings.ToLower(words[0])
	for _, w := range words[1:] {
		result += toTitleCase(w)
	}
	return result
}

func daysInYear(year int) int {
	if year%400 == 0 || (year%4 == 0 && year%100 != 0) {
		return 366
	}
	return 365
}
