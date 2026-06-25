package envtools

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// sensitiveKeywords là các từ khóa nhận diện biến môi trường nhạy cảm.
var sensitiveKeywords = []string{
	"PASSWORD", "PASSWD", "SECRET", "TOKEN", "KEY",
	"CREDENTIAL", "AUTH", "PRIVATE", "API_KEY", "APIKEY",
}

// GetEnvVar lấy giá trị của một biến môi trường cụ thể.
// Ứng dụng thực tế: Kiểm tra cấu hình runtime (DB_HOST, PORT, SERVICE_URL).
func GetEnvVar(name string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fmt.Sprintf("Biến môi trường '%s' không tồn tại", name)
	}
	if value == "" {
		return fmt.Sprintf("Biến môi trường '%s' tồn tại nhưng có giá trị rỗng", name)
	}
	// Ẩn các giá trị nhạy cảm
	if isSensitiveKey(name) {
		return fmt.Sprintf("%s = ******* (ẩn vì chứa thông tin nhạy cảm)", name)
	}
	return fmt.Sprintf("%s = %s", name, value)
}

// ListEnvVars liệt kê các biến môi trường theo prefix, ẩn giá trị nhạy cảm.
// Ứng dụng thực tế: Kiểm tra toàn bộ cấu hình môi trường khi debug deployment.
func ListEnvVars(prefix string) string {
	all := os.Environ()
	sort.Strings(all)

	prefixUpper := strings.ToUpper(prefix)
	var matches []string
	for _, entry := range all {
		if prefixUpper != "" && !strings.HasPrefix(strings.ToUpper(entry), prefixUpper) {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		if isSensitiveKey(key) {
			matches = append(matches, fmt.Sprintf("  %s=******* (ẩn)", key))
		} else {
			matches = append(matches, fmt.Sprintf("  %s=%s", key, val))
		}
	}

	if len(matches) == 0 {
		if prefix == "" {
			return "Không tìm thấy biến môi trường nào"
		}
		return fmt.Sprintf("Không tìm thấy biến môi trường nào bắt đầu bằng '%s'", prefix)
	}

	label := prefix
	if prefix == "" {
		label = "(tất cả)"
	}
	return fmt.Sprintf(
		"Biến môi trường (prefix: %s) — %d kết quả:\n%s\n%s",
		label, len(matches),
		strings.Repeat("─", 60),
		strings.Join(matches, "\n"),
	)
}

func isSensitiveKey(key string) bool {
	keyUpper := strings.ToUpper(key)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(keyUpper, kw) {
			return true
		}
	}
	return false
}
