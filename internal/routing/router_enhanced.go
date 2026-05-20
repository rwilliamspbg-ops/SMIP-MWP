// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
// Enhanced routing with LPM support and priority ranking

package routing

import (
	"bytes"
)

// LPMEntry represents a longest-prefix-match entry for routing
type LPMEntry struct {
	Prefix    []byte // Prefix bits (packed bytes)
	PrefixLen uint8
	NextHopID [32]byte
	Metric    int
}

// EnhancedTable wraps the baseline Table with optional LPM entries.
// This keeps advanced routing support additive and avoids duplicate core types.
type EnhancedTable struct {
	base       *Table
	lpmEntries []LPMEntry
}

func NewEnhancedTable(base *Table) *EnhancedTable {
	if base == nil {
		base = NewTable()
	}
	return &EnhancedTable{base: base, lpmEntries: make([]LPMEntry, 0, 16)}
}

func (t *EnhancedTable) Base() *Table {
	return t.base
}

// AddLPMRoute adds a longest-prefix-match route entry
func (t *EnhancedTable) AddLPMRoute(prefix []byte, prefixLen uint8, nextHopID [32]byte, metric int) {
	if t == nil {
		return
	}
	entry := LPMEntry{Prefix: append([]byte(nil), prefix...), PrefixLen: prefixLen, NextHopID: nextHopID, Metric: metric}
	t.lpmEntries = append(t.lpmEntries, entry)
}

// LookupLPM performs longest-prefix-match lookup
func (t *EnhancedTable) LookupLPM(dstID [32]byte) ([32]byte, int, bool) {
	if t == nil {
		return [32]byte{}, 0, false
	}
	bestLen := -1
	var best [32]byte
	for _, entry := range t.lpmEntries {
		if matchesPrefix(dstID, entry.Prefix, entry.PrefixLen) && int(entry.PrefixLen) > bestLen {
			bestLen = int(entry.PrefixLen)
			best = entry.NextHopID
		}
	}
	if bestLen < 0 {
		return [32]byte{}, 0, false
	}
	return best, bestLen, true
}

// matchesPrefix checks if dstID matches the given prefix with the given length
func matchesPrefix(dstID [32]byte, prefix []byte, prefixLen uint8) bool {
	fullBytes := int(prefixLen / 8)
	remainingBits := int(prefixLen % 8)
	if len(prefix) < fullBytes {
		return false
	}
	dst := dstID[:]
	if fullBytes > 0 && !bytes.Equal(dst[:fullBytes], prefix[:fullBytes]) {
		return false
	}
	if remainingBits == 0 {
		return true
	}
	if len(prefix) < fullBytes+1 {
		return false
	}
	mask := byte(0xFF << (8 - remainingBits))
	return (dst[fullBytes] & mask) == (prefix[fullBytes] & mask)
}
