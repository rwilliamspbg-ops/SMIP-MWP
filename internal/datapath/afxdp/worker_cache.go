// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package afxdp

// worker-local small session cache to avoid frequent map lookups on hot sessions.
// Very small fixed-size direct-mapped circular buffer; designed for simplicity
// and minimal per-packet overhead.

type sessionCacheEntry struct {
	id   [16]byte
	sess *Session
}

type workerSessionCache struct {
	entries [8]sessionCacheEntry
	next    uint8 // next insert index
	hits    uint64
	misses  uint64
}

// Get returns a cached session or nil.
func (wc *workerSessionCache) Get(id [16]byte) *Session {
	for i := 0; i < len(wc.entries); i++ {
		e := &wc.entries[i]
		if e.sess != nil && e.id == id {
			wc.hits++
			return e.sess
		}
	}
	wc.misses++
	return nil
}

// Put inserts a session into the cache, evicting the oldest entry.
func (wc *workerSessionCache) Put(id [16]byte, s *Session) {
	idx := wc.next % uint8(len(wc.entries))
	wc.entries[idx].id = id
	wc.entries[idx].sess = s
	wc.next++
}
