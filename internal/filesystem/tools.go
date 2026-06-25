package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxReadSize = 1 * 1024 * 1024 // giới hạn đọc file: 1MB

// ListDirectory liệt kê nội dung của một thư mục.
// Ứng dụng thực tế: AI agent duyệt project source code, kiểm tra cấu trúc file.
func ListDirectory(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("không thể đọc thư mục '%s': %w", path, err)
	}

	absPath, _ := filepath.Abs(path)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Thư mục: %s\n", absPath))
	sb.WriteString(strings.Repeat("─", 72) + "\n")
	sb.WriteString(fmt.Sprintf("%-6s  %-40s  %10s  %s\n", "Loại", "Tên", "Kích thước", "Sửa đổi lần cuối"))
	sb.WriteString(strings.Repeat("─", 72) + "\n")

	dirCount, fileCount := 0, 0
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		label := "[FILE]"
		if entry.IsDir() {
			label = "[DIR] "
			dirCount++
		} else {
			fileCount++
		}
		sb.WriteString(fmt.Sprintf("%-6s  %-40s  %10s  %s\n",
			label,
			entry.Name(),
			formatFileSize(info.Size()),
			info.ModTime().Format("2006-01-02 15:04:05"),
		))
	}
	sb.WriteString(strings.Repeat("─", 72) + "\n")
	sb.WriteString(fmt.Sprintf("Tổng: %d file, %d thư mục", fileCount, dirCount))
	return sb.String(), nil
}

// ReadFileContent đọc nội dung của một file văn bản (giới hạn 1MB).
// Ứng dụng thực tế: AI đọc file config, source code, log để phân tích.
func ReadFileContent(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("không tìm thấy file '%s': %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("'%s' là thư mục, không phải file", path)
	}
	if info.Size() > maxReadSize {
		return "", fmt.Errorf("file quá lớn (%s), giới hạn đọc là 1MB", formatFileSize(info.Size()))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("không thể đọc file: %w", err)
	}
	return string(data), nil
}

// WriteFileContent ghi nội dung vào một file (tạo mới nếu chưa tồn tại).
// Ứng dụng thực tế: AI tạo file config, báo cáo, code snippet tự động.
func WriteFileContent(path, content string) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("không thể tạo thư mục '%s': %w", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("không thể ghi file: %w", err)
	}
	absPath, _ := filepath.Abs(path)
	return fmt.Sprintf("Đã ghi thành công!\nFile: %s\nKích thước: %s", absPath, formatFileSize(int64(len(content)))), nil
}

// GetFileInfo trả về metadata của một file hoặc thư mục.
// Ứng dụng thực tế: Kiểm tra file tồn tại, xác minh quyền truy cập trước khi xử lý.
func GetFileInfo(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("không tìm thấy '%s': %w", path, err)
	}
	absPath, _ := filepath.Abs(path)
	label := "File"
	if info.IsDir() {
		label = "Thư mục"
	}
	return fmt.Sprintf(
		"Thông tin: %s\n%s\nĐường dẫn tuyệt đối : %s\nLoại               : %s\nKích thước         : %s\nSửa đổi lần cuối   : %s\nQuyền truy cập     : %s",
		path,
		strings.Repeat("─", 60),
		absPath,
		label,
		formatFileSize(info.Size()),
		info.ModTime().Format("2006-01-02 15:04:05"),
		info.Mode().String(),
	), nil
}

// SearchInFile tìm kiếm chuỗi văn bản trong một file và trả về các dòng khớp.
// Ứng dụng thực tế: Tìm lỗi trong log file, tìm config key trong file cấu hình.
func SearchInFile(path, pattern string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("không thể đọc file: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	patternLower := strings.ToLower(pattern)

	var matches []string
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), patternLower) {
			matches = append(matches, fmt.Sprintf("  Dòng %4d │ %s", i+1, line))
		}
	}
	if len(matches) == 0 {
		return fmt.Sprintf("Không tìm thấy '%s' trong file %s", pattern, path), nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tìm thấy %d kết quả cho '%s' trong %s:\n", len(matches), pattern, path))
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	for _, m := range matches {
		sb.WriteString(m + "\n")
	}
	return sb.String(), nil
}

func formatFileSize(size int64) string {
	switch {
	case size >= 1024*1024*1024:
		return fmt.Sprintf("%.2f GB", float64(size)/1024/1024/1024)
	case size >= 1024*1024:
		return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
	case size >= 1024:
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
