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

// getDiskSpaceForPath gets disk space for the first imported file
func (r *DefaultImportReportingSystem) getDiskSpaceForImportedFiles(importedFiles []string) int64 {
	if len(importedFiles) > 0 {
		dir := filepath.Dir(importedFiles[0])
		return getDiskSpaceAvailable(dir)
	}
	return 0
}