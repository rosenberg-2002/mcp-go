package sysinfo

import (
	"fmt"
	"runtime"
)

// GetBasicInfo trả về thông tin hệ điều hành dưới dạng chuỗi
func GetBasicInfo() string {
	return fmt.Sprintf("OS: %s\nArchitecture: %s\nLogical CPUs: %d\nGo Version: %s",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), runtime.Version())
}