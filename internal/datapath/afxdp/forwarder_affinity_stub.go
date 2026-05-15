//go:build !linux || !withafxdp
// +build !linux !withafxdp

package afxdp

// SetCurrentThreadAffinity is a no-op on unsupported builds.
func SetCurrentThreadAffinity(cpu int) error { return nil }
