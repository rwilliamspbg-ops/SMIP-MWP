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
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func randomID(t *testing.T) [32]byte {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("rand read: %v", err)
	}
	return b
}

func TestUpdateAndLookup(t *testing.T) {
	tbl := NewTable()
	dst := randomID(t)
	nh := randomID(t)
	tbl.UpdateRoute(RouteEntry{DestID: dst, NextHopID: nh})

	got, ok := tbl.LookupNextHop(dst, 0)
	if !ok {
		t.Fatalf("expected route for %s", hex.EncodeToString(dst[:4]))
	}
	if got != nh {
		t.Fatalf("unexpected next hop")
	}
}

func TestPredictive(t *testing.T) {
	tbl := NewTable()
	// Populate several entries
	for i := 0; i < 8; i++ {
		dst := randomID(t)
		nh := randomID(t)
		tbl.UpdateRoute(RouteEntry{DestID: dst, NextHopID: nh})
	}
	src := randomID(t)
	dst := randomID(t)
	nh, ok := tbl.PredictiveNextHop(src, dst, 42)
	if !ok {
		t.Fatalf("expected predictive next hop")
	}
	_ = nh // just ensure we returned something deterministic
}
