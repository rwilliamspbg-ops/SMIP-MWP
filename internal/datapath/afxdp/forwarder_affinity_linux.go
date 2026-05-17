//go:build linux && withafxdp
// +build linux,withafxdp

package afxdp

import (
	"fmt"
	"syscall"
	"unsafe"
)

// SYS_gettid is the Linux syscall number for gettid (varies by platform)
const (
	// Linux x86_64 gettid syscall
	SYS_gettid = 186 // amd64
)

// SetCurrentThreadAffinity pins the current OS thread to the given CPU using
// sched_setaffinity(2). This is a best-effort helper; CPUs >= 64 are not
// supported in this simple implementation.
func SetCurrentThreadAffinity(cpu int) error {
	if cpu < 0 {
		return fmt.Errorf("invalid cpu: %d", cpu)
	}
	// Lock OSThread must be called by caller (SpawnPerCPUWorkers does this).
	// Use gettid to address the current thread.
	tid, _, errno := syscall.RawSyscall(SYS_gettid, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("gettid failed: %v", errno)
	}

	if cpu >= 64 {
		return fmt.Errorf("cpu >=64 not supported by simple affinity helper")
	}

	var mask uint64 = 1 << uint(cpu)
	// Use sched_setaffinity(tid, sizeof(mask), &mask)
	_, _, errno = syscall.RawSyscall(syscall.SYS_SCHED_SETAFFINITY, tid, uintptr(unsafe.Sizeof(mask)), uintptr(unsafe.Pointer(&mask)))
	if errno != 0 {
		return fmt.Errorf("sched_setaffinity failed: %v", errno)
	}

	// re-check using sched_getaffinity (optional)
	return nil
}
