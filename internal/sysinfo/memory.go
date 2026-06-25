package sysinfo

import (
	"fmt"
	"runtime"
)

// GetMemoryStats trả về thông tin sử dụng RAM của tiến trình hiện tại
func GetMemoryStats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocMB := m.Alloc / 1024 / 1024
	sysMB := m.Sys / 1024 / 1024

	return fmt.Sprintf("Memory Allocated: %v MB\nTotal Memory from OS: %v MB", allocMB, sysMB)
}