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
		bsize := stat.Bsize
		if bsize < 0 {
			bsize = 0
		}
		// Safe conversion: clamp to MaxInt64 to prevent overflow
		maxBlocks := uint64(math.MaxInt64) / uint64(bsize+1)
		if bavail > maxBlocks {
			return math.MaxInt64
		}
		return int64(bavail) * int64(bsize) // #nosec G115 -- overflow prevented by clamp above
	}
	return 0
}
