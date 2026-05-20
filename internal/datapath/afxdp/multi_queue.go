// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
// Multi-queue forwarder for scalable AF_XDP forwarding

package afxdp

import (
	"context"
	"runtime"
	"sync"
)

type MultiQueueForwarder struct {
	forwarders []*Forwarder
	config     Config
	wg         sync.WaitGroup
}

func NewMultiQueueForwarder(cfg Config, rt interface{}) (*MultiQueueForwarder, error) {
	// Auto-detect NIC queue count
	numQueues := cfg.NumWorkers
	if numQueues == 0 {
		numQueues = runtime.NumCPU()
	}
	
	forwarders := make([]*Forwarder, numQueues)
	for i := 0; i < numQueues; i++ {
		fwd := &Forwarder{cfg: cfg}
		forwarders[i] = fwd
	}
	
	return &MultiQueueForwarder{forwarders: forwarders, config: cfg}, nil
}

func (m *MultiQueueForwarder) Run(ctx context.Context) error {
	for i, fwd := range m.forwarders {
		m.wg.Add(1)
		go func(id int, f *Forwarder) {
			defer m.wg.Done()
			runtime.LockOSThread() // Pin to core for each worker
			f.Run(ctx)
		}(i, fwd)
	}
	m.wg.Wait()
	return nil
}

func (m *MultiQueueForwarder) GetStats() ForwarderStats {
	var total rxPackets uint64
	for _, fwd := range m.forwarders {
		total += fwd.stats.Load()
	}
	return rxPackets: total
}
