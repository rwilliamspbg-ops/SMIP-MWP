// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package wire

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func random32(t *testing.T) [32]byte {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return b
}

func TestMarshalAndParse(t *testing.T) {
	src := random32(t)
	dst := random32(t)
	var sid [16]byte
	if _, err := rand.Read(sid[:]); err != nil {
		t.Fatalf("rand sid: %v", err)
	}

	h := Header{
		SrcID:     src,
		DstID:     dst,
		FlowLabel: 0xdeadbeef,
		SeqNum:    42,
		SessionID: sid,
		Flags:     0x1,
		Length:    128,
	}

	buf := NewHeaderBuffer(int(h.Length))
	if err := h.Marshal(buf); err != nil {
		t.Fatalf("marshal: %v", err)
	}

	parsed, err := ParseHeader(buf)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if !bytes.Equal(parsed.SrcID[:], h.SrcID[:]) {
		t.Fatalf("src mismatch")
	}
	if !bytes.Equal(parsed.DstID[:], h.DstID[:]) {
		t.Fatalf("dst mismatch")
	}
	if parsed.FlowLabel != h.FlowLabel || parsed.SeqNum != h.SeqNum || parsed.Length != h.Length {
		t.Fatalf("fields mismatch")
	}
}
