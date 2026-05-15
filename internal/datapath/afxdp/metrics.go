package afxdp

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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
	// Start background flusher to aggregate per-worker atomic counters into
	// Prometheus metrics periodically. Interval kept short for CI/test feedback.
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			flushWorkerCounters()
		}
	}()
}

var (
	rxCounters      sync.Map // map[int]*uint64
	txCounters      sync.Map
	droppedCounters sync.Map
)

func getCounter(m *sync.Map, worker int) *uint64 {
	if v, ok := m.Load(worker); ok {
		return v.(*uint64)
	}
	var z uint64
	p := &z
	actual, _ := m.LoadOrStore(worker, p)
	return actual.(*uint64)
}

func flushWorkerCounters() {
	rxCounters.Range(func(k, v interface{}) bool {
		worker := k.(int)
		p := v.(*uint64)
		val := atomic.SwapUint64(p, 0)
		if val > 0 {
			rxPacketsVec.WithLabelValues(fmt.Sprint(worker)).Add(float64(val))
			rxPackets.Add(float64(val))
		}
		return true
	})
	txCounters.Range(func(k, v interface{}) bool {
		worker := k.(int)
		p := v.(*uint64)
		val := atomic.SwapUint64(p, 0)
		if val > 0 {
			txPacketsVec.WithLabelValues(fmt.Sprint(worker)).Add(float64(val))
			txPackets.Add(float64(val))
		}
		return true
	})
	droppedCounters.Range(func(k, v interface{}) bool {
		worker := k.(int)
		p := v.(*uint64)
		val := atomic.SwapUint64(p, 0)
		if val > 0 {
			droppedPacketsVec.WithLabelValues(fmt.Sprint(worker)).Add(float64(val))
			droppedPackets.Add(float64(val))
		}
		return true
	})
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
	p := getCounter(&rxCounters, worker)
	atomic.AddUint64(p, uint64(n))
}

func IncTxWorker(worker int, n int) {
	if n <= 0 {
		return
	}
	p := getCounter(&txCounters, worker)
	atomic.AddUint64(p, uint64(n))
}

func IncDroppedWorker(worker int, n int) {
	if n <= 0 {
		return
	}
	p := getCounter(&droppedCounters, worker)
	atomic.AddUint64(p, uint64(n))
}

func ObserveProcessingLatency(worker int, seconds float64) {
	processingLatency.WithLabelValues(fmt.Sprint(worker)).Observe(seconds)
}
