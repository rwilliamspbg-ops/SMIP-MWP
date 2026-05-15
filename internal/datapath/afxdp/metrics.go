package afxdp

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	rxPackets = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "smip_mwp",
		Subsystem: "afxdp",
		Name:      "rx_packets_total",
		Help:      "Total received packets",
	})
	txPackets = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "smip_mwp",
		Subsystem: "afxdp",
		Name:      "tx_packets_total",
		Help:      "Total transmitted packets",
	})
	droppedPackets = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "smip_mwp",
		Subsystem: "afxdp",
		Name:      "dropped_packets_total",
		Help:      "Total dropped/failed-to-process packets",
	})
	handshakeCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "smip_mwp",
		Subsystem: "crypto",
		Name:      "handshakes_total",
		Help:      "Total completed handshake/session establishments",
	})
	cryptoErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "smip_mwp",
		Subsystem: "crypto",
		Name:      "errors_total",
		Help:      "Total cryptographic errors (encrypt/decrypt failures)",
	})
)

func init() {
	prometheus.MustRegister(rxPackets, txPackets, droppedPackets, handshakeCount, cryptoErrors)
}

func IncRx(n int) {
	if n > 0 {
		rxPackets.Add(float64(n))
	}
}
func IncTx(n int) {
	if n > 0 {
		txPackets.Add(float64(n))
	}
}
func IncDropped(n int) {
	if n > 0 {
		droppedPackets.Add(float64(n))
	}
}
func IncHandshake()   { handshakeCount.Inc() }
func IncCryptoError() { cryptoErrors.Inc() }
