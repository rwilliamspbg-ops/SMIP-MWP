//go:build !withafxdp
// +build !withafxdp

package afxdp

import (
	"log"
	"os"
	"sync"

	"smip-mwp/internal/routing"
)

// NewForwarder creates the non-AF_XDP stub forwarder used for unit tests and
// development when the withafxdp build tag is not set.
func NewForwarder(cfg Config, routeTable *routing.Table) (*Forwarder, error) {
	l := log.New(os.Stdout, "afxdp: ", log.LstdFlags)
	f := &Forwarder{cfg: cfg, logger: l, routeTable: routeTable, sessions: make(map[[16]byte]*Session)}
	// initialize pktPool sized to FrameSize (fallback to 2048 if unset)
	size := cfg.FrameSize
	if size <= 0 {
		size = 2048
	}
	f.pktPool = &sync.Pool{New: func() interface{} { return make([]byte, size) }}
	f.logger.Printf("stub forwarder created iface=%s", cfg.Interface)
	return f, nil
}
