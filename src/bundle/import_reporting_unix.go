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
		bavail := stat.Bavail
		if bavail > uint64(math.MaxInt64) {
			bavail = uint64(math.MaxInt64)
		}
		return int64(bavail) * int64(stat.Bsize) // #nosec G115 -- bavail bounds checked above; Bsize is int32 so int64 widening is safe
	}
	return 0
}
