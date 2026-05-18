package afxdp

import (
	"sync"
	"testing"
)

// BenchmarkPacketAllocate measures allocating packets with make() each time.
func BenchmarkPacketAllocate(b *testing.B) {
	size := 2048
	for i := 0; i < b.N; i++ {
		pkt := make([]byte, size)
		_ = pkt
	}
}

// BenchmarkPacketPool measures using a sync.Pool to reuse packet buffers.
func BenchmarkPacketPool(b *testing.B) {
	size := 2048
	pool := &sync.Pool{New: func() interface{} { buf := make([]byte, size); return &buf }}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pktPtr := pool.Get().(*[]byte)
		pkt := *pktPtr
		// use packet
		_ = pkt
		pool.Put(pktPtr)
	}
}
