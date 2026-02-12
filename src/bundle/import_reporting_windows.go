//go:build windows

package bundle

import (
	"syscall"
	"unsafe"
)

// getDiskSpaceAvailable gets available disk space for Windows systems
func getDiskSpaceAvailable(dir string) int64 {
	if dir == "" {
		return 0
	}

	// Simple Windows implementation
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return 0
	}
	defer kernel32.Release()

	getDiskFreeSpaceEx, err := kernel32.FindProc("GetDiskFreeSpaceExW")
	if err != nil {
		return 0
	}

	// Convert to UTF-16
	dirPtr, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return 0
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	ret, _, _ := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(dirPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return 0
	}

	return int64(freeBytesAvailable)
}

