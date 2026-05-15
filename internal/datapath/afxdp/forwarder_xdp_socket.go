//go:build withafxdp
// +build withafxdp

package afxdp

import (
	"fmt"

	"reflect"

	xdp "github.com/asavie/xdp"
)

// XDPSocket wraps an asavie/xdp socket bound to a network interface and queue.
type XDPSocket struct {
	s *xdp.Socket
}

// NewXDPSocket creates and binds an AF_XDP socket to iface:queue using the
// provided UMEM. This implementation uses the asavie/xdp library; adjust the
// call sites if the library API changes.
func NewXDPSocket(iface string, queue int, umem *UMEM) (*XDPSocket, error) {
	if iface == "" {
		return nil, fmt.Errorf("iface required")
	}
	if umem == nil || umem.u == nil {
		return nil, fmt.Errorf("umem required")
	}

	// Create socket config
	cfg := xdp.SocketConfig{
		Iface: iface,
		Queue: uint32(queue),
	}
	sock, err := xdp.NewSocket(&cfg, umem.u)
	if err != nil {
		return nil, fmt.Errorf("xdp.NewSocket: %w", err)
	}
	return &XDPSocket{s: sock}, nil
}

// Close releases socket resources.
func (s *XDPSocket) Close() error {
	if s == nil || s.s == nil {
		return nil
	}
	return s.s.Close()
}

// Poll attempts to receive up to `max` frames from the underlying xdp.Socket.
// This method uses reflection to call the available batch/recv API on the
// asavie/xdp Socket implementation to remain compatible with minor API
// variations between versions. It returns a slice of byte slices (one per
// received frame).
func (s *XDPSocket) Poll(max int) ([][]byte, error) {
	if s == nil || s.s == nil {
		return nil, fmt.Errorf("socket not initialized")
	}
	v := reflect.ValueOf(s.s)
	// Candidate method names in order of preference.
	candidates := []string{"RecvBatch", "ReceiveBatch", "ReadBatch", "Recv", "Receive", "Read"}
	for _, name := range candidates {
		m := v.MethodByName(name)
		if !m.IsValid() {
			continue
		}
		// Prepare args: if method accepts an int, pass max; otherwise none.
		var args []reflect.Value
		if m.Type().NumIn() == 1 {
			// Support int or uint32 param
			paramType := m.Type().In(0).Kind()
			if paramType == reflect.Int {
				args = []reflect.Value{reflect.ValueOf(max)}
			} else if paramType == reflect.Uint32 {
				args = []reflect.Value{reflect.ValueOf(uint32(max))}
			} else {
				// Unsupported param type; skip
				continue
			}
		}
		rets := m.Call(args)
		// Expect ([][]byte, error) or ([]byte, error) or (int, [][]byte, error)
		if len(rets) == 2 {
			errVal := rets[1]
			var err error
			if !errVal.IsNil() {
				err = errVal.Interface().(error)
			}
			first := rets[0]
			// If first is [][]byte
			if first.Kind() == reflect.Slice {
				// Check element kind
				elem := first.Type().Elem()
				if elem.Kind() == reflect.Slice {
					// It's [][]byte
					out := make([][]byte, first.Len())
					for i := 0; i < first.Len(); i++ {
						out[i] = first.Index(i).Interface().([]byte)
					}
					return out, err
				}
				// If it's []byte (single frame), wrap
				if elem.Kind() == reflect.Uint8 {
					b := first.Interface().([]byte)
					return [][]byte{b}, err
				}
			}
		} else if len(rets) == 3 {
			// Possible signature: (int, [][]byte, error)
			errVal := rets[2]
			var err error
			if !errVal.IsNil() {
				err = errVal.Interface().(error)
			}
			second := rets[1]
			if second.Kind() == reflect.Slice && second.Type().Elem().Kind() == reflect.Slice {
				out := make([][]byte, second.Len())
				for i := 0; i < second.Len(); i++ {
					out[i] = second.Index(i).Interface().([]byte)
				}
				return out, err
			}
		}
		// If we reach here, method returned an unexpected shape; continue to next
	}
	return nil, fmt.Errorf("no compatible recv API found on xdp.Socket")
}

// Send attempts to transmit a batch of packets using the underlying xdp.Socket.
// It uses reflection to call common transmit APIs (SendBatch, Transmit, Write,
// Tx) on the Socket. It returns an error if no compatible API is found.
func (s *XDPSocket) Send(pkts [][]byte) error {
	if s == nil || s.s == nil {
		return fmt.Errorf("socket not initialized")
	}
	v := reflect.ValueOf(s.s)
	candidates := []string{"SendBatch", "Transmit", "Write", "Tx"}
	for _, name := range candidates {
		m := v.MethodByName(name)
		if !m.IsValid() {
			continue
		}
		// Determine whether method accepts [][]byte or other shapes
		// We attempt to pass [][]byte directly if the method expects a slice.
		if m.Type().NumIn() == 1 {
			inTy := m.Type().In(0)
			if inTy.Kind() == reflect.Slice {
				// Build reflect value for pkts
				vpkts := reflect.ValueOf(pkts)
				rets := m.Call([]reflect.Value{vpkts})
				// If method returns error
				if len(rets) == 1 {
					if rets[0].IsNil() {
						return nil
					}
					return rets[0].Interface().(error)
				}
				if len(rets) == 2 {
					// e.g., (int, error)
					if !rets[1].IsNil() {
						return rets[1].Interface().(error)
					}
					return nil
				}
				return nil
			}
		}
	}
	return fmt.Errorf("no compatible send API found on xdp.Socket")
}
