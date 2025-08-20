//go:build unix

package bundle

import (
	"syscall"
)

// getDiskSpaceAvailable gets available disk space for Unix systems
func getDiskSpaceAvailable(dir string) int64 {
	if dir == "" {
		return 0
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err == nil {
		return int64(stat.Bavail) * int64(stat.Bsize)
	}
	return 0
}
