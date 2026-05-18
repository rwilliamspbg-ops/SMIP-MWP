// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package routing

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// RouteEntry represents a routing decision for a sovereign destination.
type RouteEntry struct {
	DestID     [32]byte
	NextHopID  [32]byte
	NextHopMAC net.HardwareAddr
	Metric     int
	LastSeen   time.Time
}

// Table is a simple in-memory sovereign routing table with a small predictive
// next-hop selector used as a stub for ML/policy-driven decisions.
type Table struct {
	mu      sync.RWMutex
	entries map[[32]byte]RouteEntry
}

// NewTable creates an empty routing table.
func NewTable() *Table {
	return &Table{entries: make(map[[32]byte]RouteEntry)}
}

// UpdateRoute inserts or refreshes a route entry.
func (t *Table) UpdateRoute(e RouteEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e.LastSeen = time.Now()
	t.entries[e.DestID] = e
}

// RemoveRoute removes a route by destination ID.
func (t *Table) RemoveRoute(dest [32]byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, dest)
}

// LookupNextHop returns an exact-match next-hop for dstID if present.
func (t *Table) LookupNextHop(dstID [32]byte, flowLabel uint32) (nextHopID [32]byte, ok bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if e, found := t.entries[dstID]; found {
		return e.NextHopID, true
	}
	return [32]byte{}, false
}

// PredictiveNextHop returns a deterministic, lightweight predictive choice when
// there is no exact route. It hashes the src/dst/flowLabel and picks an entry
// from the table for a best-effort next hop. This is intentionally simple and
// meant as a pluggable hook for more advanced predictors.
func (t *Table) PredictiveNextHop(srcID, dstID [32]byte, flowLabel uint32) (nextHopID [32]byte, ok bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.entries) == 0 {
		return [32]byte{}, false
	}
	// Prefer exact match
	if e, found := t.entries[dstID]; found {
		return e.NextHopID, true
	}
	// Build a stable list of keys
	keys := make([][32]byte, 0, len(t.entries))
	for k := range t.entries {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return string(keys[i][:]) < string(keys[j][:]) })

	h := sha256.New()
	h.Write(srcID[:])
	h.Write(dstID[:])
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], flowLabel)
	h.Write(b[:])
	sum := h.Sum(nil)
	idx := binary.BigEndian.Uint32(sum[:4]) % uint32(len(keys))
	chosen := t.entries[keys[idx]]
	return chosen.NextHopID, true
}

// LookupOrPredict returns an exact next hop if present, otherwise a predictive
// next hop. The boolean is true when any next hop is returned.
func (t *Table) LookupOrPredict(srcID, dstID [32]byte, flowLabel uint32) (nextHopID [32]byte, ok bool) {
	if nh, found := t.LookupNextHop(dstID, flowLabel); found {
		return nh, true
	}
	return t.PredictiveNextHop(srcID, dstID, flowLabel)
}

// RoutePolicy defines the action taken when traffic matches a policy rule.
type RoutePolicy struct {
	NextHopID [32]byte // The sovereign Crypto ID of the next physical hop/endpoint
	QueueID   int      // Which AF_XDP queue should process this flow (for steering)
	Priority  int      // Lower number = higher priority match
}

// Router is a simple policy manager that complements Table. It provides a
// higher-level policy lookup path and seeded defaults for testing and demo.
type Router struct {
	sync.RWMutex
	policies map[uint64]RoutePolicy // Key: Hash(SrcID, DstID, FlowLabel)
}

// NewRouter initializes a router with default/pre-defined policies.
func NewRouter() *Router {
	r := &Router{
		policies: make(map[uint64]RoutePolicy),
	}
	r.SeedDefaultPolicies()
	return r
}

// SeedDefaultPolicies loads foundational, trusted routing policies.
func (r *Router) SeedDefaultPolicies() {
	r.Lock()
	defer r.Unlock()

	defaultPolicy := RoutePolicy{
		NextHopID: [32]byte{},
		QueueID:   0,
		Priority:  10,
	}
	r.policies[0] = defaultPolicy
}

// computeFlowKey hashes the key components of the packet header for lookup.
func (r *Router) computeFlowKey(srcID [32]byte, dstID [32]byte, flowLabel uint32) uint64 {
	var key uint64
	for i := 0; i < 8; i++ {
		key ^= (uint64(srcID[i*4])<<32 | uint64(dstID[i*4]))
	}
	key ^= uint64(flowLabel)
	return key
}

// LookupPolicy attempts to find the best policy for a given packet header.
func (r *Router) LookupPolicy(srcID [32]byte, dstID [32]byte, flowLabel uint32) (RoutePolicy, error) {
	key := r.computeFlowKey(srcID, dstID, flowLabel)

	r.RLock()
	if policy, ok := r.policies[key]; ok {
		r.RUnlock()
		return policy, nil
	}
	r.RUnlock()

	r.RLock()
	defer r.RUnlock()
	if policy, ok := r.policies[0]; ok {
		return policy, nil
	}
	return RoutePolicy{}, fmt.Errorf("no policy available")
}

// UpdatePolicy allows external agents (like the consensus plane) to dynamically update the table.
func (r *Router) UpdatePolicy(srcID [32]byte, dstID [32]byte, flowLabel uint32, nextHopID [32]byte, queueID int) {
	key := r.computeFlowKey(srcID, dstID, flowLabel)

	r.Lock()
	defer r.Unlock()

	r.policies[key] = RoutePolicy{
		NextHopID: nextHopID,
		QueueID:   queueID,
		Priority:  1, // Highest priority update
	}
	fmt.Printf("SUCCESS: Policy updated for key %x -> Queue %d\n", key, queueID)
}
