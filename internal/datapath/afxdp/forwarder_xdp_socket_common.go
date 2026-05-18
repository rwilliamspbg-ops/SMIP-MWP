// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package afxdp

// socketBackend abstracts the backing socket implementation used by XDPSocket.
// Implementations include the mock `xdpSocketImpl` (non-asavie) and the
// `realSocketImpl` (asavie).
type socketBackend interface {
	NumFreeFillSlots() int
	GetDescs(n int) []*XDPDescriptor
	Fill(descs []*XDPDescriptor)
	Poll(maxEvents int) (int, int, error)
	Receive(n int) []*XDPDescriptor
	GetFrame(d *XDPDescriptor) []byte
	Complete(n int)
	Transmit(descs []*XDPDescriptor)
	Close() error
}

// XDPSocket wraps a backend implementation; it delegates lifecycle calls to
// the backend and provides a stable type for the forwarder code to reference.
type XDPSocket struct {
	s socketBackend
}

// Close releases socket resources
func (s *XDPSocket) Close() error {
	if s == nil || s.s == nil {
		return nil
	}
	return s.s.Close()
}
