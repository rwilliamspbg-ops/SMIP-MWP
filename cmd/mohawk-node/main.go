package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"smip-mwp/internal/routing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	iface := flag.String("iface", "eth0", "network interface")
	dry := flag.Bool("dry-run", true, "don't initialize AF_XDP; show plan only")
	metricsAddr := flag.String("metrics-addr", ":9090", "address to expose Prometheus metrics (empty to disable)")
	flag.Parse()

	rt := routing.NewTable()
	// Prime with a sample route (zeroed IDs for example purposes)
	var dest, next [32]byte
	copy(dest[:], []byte("example-dst-000000000000000000000"))
	copy(next[:], []byte("example-next-000000000000000000000"))
	rt.UpdateRoute(routing.RouteEntry{DestID: dest, NextHopID: next})

	fmt.Println("Routing table primed with sample entry")
	if *metricsAddr != "" {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			log.Printf("metrics: listening on %s\n", *metricsAddr)
			if err := http.ListenAndServe(*metricsAddr, nil); err != nil {
				log.Printf("metrics endpoint failed: %v", err)
			}
		}()
	}
	if *dry {
		fmt.Printf("Dry run: would create AF_XDP forwarder on %s\n", *iface)
		fmt.Println("To run with AF_XDP, rebuild with -tags=withafxdp and ensure kernel libs are installed:")
		fmt.Println("  go run -tags=withafxdp ./cmd/mohawk-node --iface=eth0")
		return
	}

	fmt.Println("Dry-run disabled but AF_XDP support not compiled in. Rebuild with -tags=withafxdp.")
}
