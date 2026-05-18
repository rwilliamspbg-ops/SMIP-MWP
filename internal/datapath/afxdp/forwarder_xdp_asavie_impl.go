//go:build withafxdp && asavie
// +build withafxdp,asavie

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

	xdpPkg "github.com/asavie/xdp"
	"github.com/vishvananda/netlink"
)

// realSocketImpl wraps github.com/asavie/xdp.Socket and adapts it to the
// socketBackend interface used by the forwarder.
type realSocketImpl struct {
	xsk       *xdpPkg.Socket
	frameSize int
}

func (r *realSocketImpl) NumFreeFillSlots() int {
	return r.xsk.NumFreeFillSlots()
}

func (r *realSocketImpl) GetDescs(n int) []*XDPDescriptor {
	descs := r.xsk.GetDescs(n)
	out := make([]*XDPDescriptor, 0, len(descs))
	for _, d := range descs {
		idx := d.Addr / uint64(r.frameSize)
		out = append(out, &XDPDescriptor{Addr: idx, Len: d.Len})
	}
	return out
}

func (r *realSocketImpl) Fill(descs []*XDPDescriptor) {
	if len(descs) == 0 {
		return
	}
	out := make([]xdpPkg.Desc, 0, len(descs))
	for _, d := range descs {
		out = append(out, xdpPkg.Desc{Addr: d.Addr * uint64(r.frameSize), Len: d.Len})
	}
	r.xsk.Fill(out)
}

func (r *realSocketImpl) Poll(maxEvents int) (int, int, error) {
	return r.xsk.Poll(maxEvents)
}

func (r *realSocketImpl) Receive(n int) []*XDPDescriptor {
	descs := r.xsk.Receive(n)
	out := make([]*XDPDescriptor, 0, len(descs))
	for _, d := range descs {
		idx := d.Addr / uint64(r.frameSize)
		out = append(out, &XDPDescriptor{Addr: idx, Len: d.Len})
	}
	return out
}

func (r *realSocketImpl) GetFrame(d *XDPDescriptor) []byte {
	if d == nil {
		return nil
	}
	desc := xdpPkg.Desc{Addr: d.Addr * uint64(r.frameSize), Len: d.Len}
	return r.xsk.GetFrame(desc)
}

func (r *realSocketImpl) Complete(n int) {
	r.xsk.Complete(n)
}

func (r *realSocketImpl) Transmit(descs []*XDPDescriptor) {
	if len(descs) == 0 {
		return
	}
	out := make([]xdpPkg.Desc, 0, len(descs))
	for _, d := range descs {
		out = append(out, xdpPkg.Desc{Addr: d.Addr * uint64(r.frameSize), Len: d.Len})
	}
	r.xsk.Transmit(out)
}

func (r *realSocketImpl) Close() error {
	return r.xsk.Close()
}

// NewXDPSocket creates a real AF_XDP socket backed by asavie/xdp.Socket.
func NewXDPSocket(iface string, queue int, umem *UMEM) (*XDPSocket, error) {
	if iface == "" {
		return nil, fmt.Errorf("iface required")
	}
	// Resolve interface name to index
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to find interface %s: %w", iface, err)
	}

	// Build socket options from UMEM sizing
	var opts *xdpPkg.SocketOptions
	if umem != nil {
		opts = &xdpPkg.SocketOptions{NumFrames: umem.numFrames, FrameSize: umem.frameSize}
	}

	xsk, err := xdpPkg.NewSocket(link.Attrs().Index, queue, opts)
	if err != nil {
		hint := "" +
			"possible causes: install libbpf-dev and bpftool, ensure kernel and NIC driver support AF_XDP sockets, " +
			"and verify UMEM parameters (frame size / num frames).\n" +
			"Run: 'sudo apt install libbpf-dev bpftool ethtool', then 'ethtool -i " + iface + "', 'sudo bpftool net', and 'dmesg | tail -n 50' to gather kernel messages."
		return nil, fmt.Errorf("xdp.NewSocket failed: %w\nHint: %s", err, hint)
	}

	backend := &realSocketImpl{xsk: xsk, frameSize: optsOrDefaultFrameSize(opts)}
	return &XDPSocket{s: backend}, nil
}

func optsOrDefaultFrameSize(opts *xdpPkg.SocketOptions) int {
	if opts == nil || opts.FrameSize == 0 {
		return xdpPkg.DefaultSocketOptions.FrameSize
	}
	return opts.FrameSize
}
