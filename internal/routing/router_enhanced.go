// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
// Enhanced routing with LPM support and priority ranking

package routing

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"time"
)

// RouteEntry represents a routing decision for a sovereign destination
type RouteEntry struct {
	DestID      [32]byte
	NextHopID   [32]byte
	NextHopMAC  [6]byte
	Metric      int
	LastSeen    time.Time
	PrefixLen   uint8 // For LPM support
}

// LPMEntry represents a longest-prefix-match entry for routing
type LPMEntry struct {
	Prefix        []byte      // CIDR-style prefix (e.g., "10.0.0./24")
	PrefixLen     uint8       // Prefix length (e.g., 24 for /24)
	NextHopID     [32]byte
	Metric        int
}

// Table is an enhanced routing table with LPM support
type Table struct {
	mu           sync.RWMutex
	exactEntries map[[32]byte]RouteEntry // Exact-match entries
	lpmEntries   map[int][]LPMEntry      // LPM entries by prefix length bucket (for faster lookup)
}

func NewEnhancedTable() *Table {
	return &Table{
		exactEntries: make(map[[32]byte]RouteEntry),
		lpmEntries:   make(map[int][]LPMEntry),
	}
}

// UpdateRoute inserts or refreshes an exact-match route entry
func (t *Table) UpdateRoute(e RouteEntry) {
	t.mu.Lock()
	defer t.Unlock()
	e.LastSeen = time.Now()
	t.exactEntries[e.DestID] = e
}

// RemoveRoute removes a route by destination ID
func (t *Table) RemoveRoute(dest [32]byte) {
	t.mu.Lock()
	defer t.Unlock()
	delete(t.exactEntries, dest)
}

// LookupNextHop returns an exact-match next-hop for dstID if present
func (t *Table) LookupNextHop(dstID [32]byte, flowLabel uint32) ([32]byte, bool) {
	t.mu.RLock()
	defer t.RUnlock()
	if e, found := t.exactEntries[dstID]; found {
		return e.NextHopID, true
	}
	return [32]byte{}, false
}

// LookupOrPredict returns an exact next hop if present, otherwise a predictive next hop
func (t *Table) LookupOrPredict(srcID, dstID [32]byte, flowLabel uint32) (nextHopID [32]byte, ok bool) {
	if nh, found := t.LookupNextHop(dstID, flowLabel); found {
		return nh, true
	}
	return t.PredictiveNextHop(srcID, dstID, flowLabel)
}

// AddLPMRoute adds a longest-prefix-match route entry
func (t *Table) AddLPMRoute(prefix []byte, prefixLen uint8, nextHopID [32]byte, metric int) {
	t.mu.Lock()
	defer t.Unlock()
	
	entry := LPMEntry{
		Prefix:     prefix,
		PrefixLen:  prefixLen,
		NextHopID:  nextHopID,
		Metric:     metric,
	}
	
	bucket := int(prefixLen) % 16 // Bucket by prefix length for O(1) access
	if _, exists := t.lpmEntries[bucket]; !exists {
		t.lpmEntries[bucket] = make([]LPMEntry, 0)
	}
	t.lpmEntries[bucket] = append(t.lpmEntries[bucket], entry)
}

// LookupLPM performs longest-prefix-match lookup
func (t *Table) LookupLPM(dstID [32]byte) ([32]byte, int, bool) {
	t.mu.RLock()
	defer t.RUnlock()
	
	// Start with longest prefix length (most specific)
	bestMatch := -1
	var bestEntry LPMEntry
	
	for len := 16; len >= 0; len-- {
		bucket := len % 16
		if entries, ok := t.lpmEntries[bucket]; ok {
			for _, entry := range entries {
				// Check if dstID matches this prefix (simplified check)
				if matchesPrefix(dstID, entry.Prefix, entry.PrefixLen) {
					bestMatch = int(entry.PrefixLen)
					bestEntry = entry
					break
				}
			}
		}
		if bestMatch != -1 {
			break
		}
	}
	
	if bestMatch != -1 {
		return bestEntry.NextHopID, bestMatch, true
	}
	return [32]byte{}, 0, false
}

// matchesPrefix checks if dstID matches the given prefix with the given length
func matchesPrefix(dstID [32]byte, prefix []byte, prefixLen uint8) bool {
	if len(prefix) < int(prefixLen/8) {
		return false
	}
	
	match := true
	for i := range int(prefixLen)/8 {
		if dstID[i*4] != prefix[i] ||
		   dstID[i*4+1] != prefix[i+1] ||
		   dstID[i*4+2] != prefix[i+2] ||
		   dstID[i*4+3] != prefix[i+3] {
			match = false
			break
		}
	}
	return match
}

// PredictiveNextHop returns a deterministic, lightweight predictive choice
func (t *Table) PredictiveNextHop(srcID, dstID [32]byte, flowLabel uint32) (nextHopID [32]byte, ok bool) {
	t.mu.RLock()
	defer t.RUnlock()
	if len(t.exactEntries) == 0 {
		return [32]byte{}, false
	}
	
	keys := make([][32]byte, 0, len(t.exactEntries))
	for k := range t.exactEntries {
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
	chosen := t.exactEntries[keys[idx]]
	return chosen.NextHopID, true
}
