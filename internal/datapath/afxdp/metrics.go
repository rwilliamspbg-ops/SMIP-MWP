package afxdp

import (
	"fmt"

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
	rxPacketsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "smip_mwp",
		Subsystem: "afxdp",
		Name:      "rx_packets",
		Help:      "Received packets by worker",
	}, []string{"worker"})
	txPacketsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "smip_mwp",
		Subsystem: "afxdp",
		Name:      "tx_packets",
		Help:      "Transmitted packets by worker",
	}, []string{"worker"})
	droppedPacketsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "smip_mwp",
		Subsystem: "afxdp",
		Name:      "dropped_packets",
		Help:      "Dropped packets by worker",
	}, []string{"worker"})
	processingLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "smip_mwp",
		Subsystem: "afxdp",
		Name:      "processing_latency_seconds",
		Help:      "Per-worker packet batch processing latency in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"worker"})
)

func init() {
	prometheus.MustRegister(rxPackets, txPackets, droppedPackets, handshakeCount, cryptoErrors)
	prometheus.MustRegister(rxPacketsVec, txPacketsVec, droppedPacketsVec)
	prometheus.MustRegister(processingLatency)
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

func IncRxWorker(worker int, n int) {
	if n <= 0 {
		return
	}
	rxPacketsVec.WithLabelValues(fmt.Sprint(worker)).Add(float64(n))
	IncRx(n)
}

func IncTxWorker(worker int, n int) {
	if n <= 0 {
		return
	}
	txPacketsVec.WithLabelValues(fmt.Sprint(worker)).Add(float64(n))
	IncTx(n)
}

func IncDroppedWorker(worker int, n int) {
	if n <= 0 {
		return
	}
	droppedPacketsVec.WithLabelValues(fmt.Sprint(worker)).Add(float64(n))
	IncDropped(n)
}

func ObserveProcessingLatency(worker int, seconds float64) {
	processingLatency.WithLabelValues(fmt.Sprint(worker)).Observe(seconds)
}
