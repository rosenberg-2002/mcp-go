package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxReadSize = 1 * 1024 * 1024 // file read limit: 1 MB

// ListDirectory lists the contents of a directory.
// Practical uses: AI agents browsing project source code, checking file structure.
func ListDirectory(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("cannot read directory '%s': %w", path, err)
	}

	absPath, _ := filepath.Abs(path)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Directory: %s\n", absPath))
	sb.WriteString(strings.Repeat("─", 72) + "\n")
	sb.WriteString(fmt.Sprintf("%-6s  %-40s  %10s  %s\n", "Type", "Name", "Size", "Last Modified"))
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
	sb.WriteString(fmt.Sprintf("Total: %d file(s), %d director(ies)", fileCount, dirCount))
	return sb.String(), nil
}

// ReadFileContent reads the content of a text file (limit: 1 MB).
// Practical uses: AI reading config files, source code, logs for analysis.
func ReadFileContent(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("file not found '%s': %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("'%s' is a directory, not a file", path)
	}
	if info.Size() > maxReadSize {
		return "", fmt.Errorf("file too large (%s), read limit is 1 MB", formatFileSize(info.Size()))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}
	return string(data), nil
}

// WriteFileContent writes content to a file (creates it if it does not exist).
// Practical uses: AI auto-generating config files, reports, and code snippets.
func WriteFileContent(path, content string) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create directory '%s': %w", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("cannot write file: %w", err)
	}
	absPath, _ := filepath.Abs(path)
	return fmt.Sprintf("Successfully written!\nFile: %s\nSize: %s", absPath, formatFileSize(int64(len(content)))), nil
}

// GetFileInfo returns metadata for a file or directory.
// Practical uses: checking file existence, verifying permissions before processing.
func GetFileInfo(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("not found '%s': %w", path, err)
	}
	absPath, _ := filepath.Abs(path)
	label := "File"
	if info.IsDir() {
		label = "Directory"
	}
	return fmt.Sprintf(
		"Info: %s\n%s\nAbsolute path  : %s\nType           : %s\nSize           : %s\nLast modified  : %s\nPermissions    : %s",
		path,
		strings.Repeat("─", 60),
		absPath,
		label,
		formatFileSize(info.Size()),
		info.ModTime().Format("2006-01-02 15:04:05"),
		info.Mode().String(),
	), nil
}

// SearchInFile searches for a text string in a file and returns matching lines.
// Practical uses: finding errors in log files, locating config keys in config files.
func SearchInFile(path, pattern string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	patternLower := strings.ToLower(pattern)

	var matches []string
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), patternLower) {
			matches = append(matches, fmt.Sprintf("  Line %4d │ %s", i+1, line))
		}
	}
	if len(matches) == 0 {
		return fmt.Sprintf("No matches for '%s' in file %s", pattern, path), nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d match(es) for '%s' in %s:\n", len(matches), pattern, path))
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
