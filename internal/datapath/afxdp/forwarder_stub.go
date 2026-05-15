//go:build !withafxdp
// +build !withafxdp

package afxdp

import (
	"log"
	"os"

	"smip-mwp/internal/routing"
)

// NewForwarder creates the non-AF_XDP stub forwarder used for unit tests and
// development when the withafxdp build tag is not set.
func NewForwarder(cfg Config, routeTable *routing.Table) (*Forwarder, error) {
	l := log.New(os.Stdout, "afxdp: ", log.LstdFlags)
	f := &Forwarder{cfg: cfg, logger: l, routeTable: routeTable, sessions: make(map[[16]byte]*Session)}
	f.logger.Printf("stub forwarder created iface=%s", cfg.Interface)
	return f, nil
}
