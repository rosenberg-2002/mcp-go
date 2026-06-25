package sysinfo

import (
	"fmt"
	"runtime"
)

// GetBasicInfo returns operating system information as a string.
func GetBasicInfo() string {
	return fmt.Sprintf("OS: %s\nArchitecture: %s\nLogical CPUs: %d\nGo Version: %s",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), runtime.Version())
}
