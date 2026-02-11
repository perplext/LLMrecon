//go:build unix

package bundle

import (
	"math"
	"syscall"
)

// getDiskSpaceAvailable gets available disk space for Unix systems
func getDiskSpaceAvailable(dir string) int64 {
	if dir == "" {
		return 0
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err == nil {
		bavail := min(stat.Bavail, uint64(math.MaxInt64))
		bsize := stat.Bsize
		if bsize < 0 {
			bsize = 0
		}
		return int64(bavail) * int64(bsize)
	}
	return 0
}
