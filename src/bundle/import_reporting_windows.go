//go:build windows

package bundle

import (
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpace = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// getDiskSpaceAvailable gets available disk space for Windows systems
func getDiskSpaceAvailable(dir string) int64 {
	if dir == "" {
		return 0
	}
	
	// Convert to UTF-16 for Windows API
	dirPtr, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return 0
	}
	
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	
	ret, _, _ := getDiskFreeSpace.Call(
		uintptr(unsafe.Pointer(dirPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	
	if ret == 0 {
		return 0
	}
	
	return int64(freeBytesAvailable)

// getDiskSpaceForPath gets disk space for the first imported file
func (r *DefaultImportReportingSystem) getDiskSpaceForImportedFiles(importedFiles []string) int64 {
	if len(importedFiles) > 0 {
		dir := filepath.Dir(importedFiles[0])
		return getDiskSpaceAvailable(dir)
	}
