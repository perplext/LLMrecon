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
		// Explicit bounds check: bavail is guaranteed <= maxBlocks <= MaxInt64
		if bavail > uint64(math.MaxInt64) {
			return math.MaxInt64
		}
		safeBavail := int64(bavail)
		safeBsize := int64(bsize)
		return safeBavail * safeBsize
	}
	return 0
}
