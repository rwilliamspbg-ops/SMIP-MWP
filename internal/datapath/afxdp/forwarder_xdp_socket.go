//go:build withafxdp && !asavie
// +build withafxdp,!asavie

// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2026 rwilliamspbg-ops
//
// This file is part of SMIP-MWP.
// SMIP-MWP is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
// See the LICENSE file in the project root for details.

package afxdp

import (
	"fmt"
)

// NewXDPSocket creates an AF_XDP socket bound to iface:queue.
// In production, this would use github.com/asavie/xdp library.
func NewXDPSocket(iface string, queue int, umem *UMEM) (*XDPSocket, error) {
	if iface == "" {
		return nil, fmt.Errorf("iface required")
	}
	if umem == nil {
		return nil, fmt.Errorf("umem required")
	}

	// Create mock socket backed by provided UMEM (simulates descriptors referencing UMEM frames)
	sock := &XDPSocket{
		s: &xdpSocketImpl{
			umem:     umem,
			fillRing: make([]*XDPDescriptor, 0),
			rxRing:   make([]*XDPDescriptor, 0),
			txRing:   make([]*XDPDescriptor, 0),
			compRing: make([]*XDPDescriptor, 0),
		},
	}
	return sock, nil
}
