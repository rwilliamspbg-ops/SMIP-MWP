//! Ring mmap abstractions for AF_XDP
//!
//! These types provide lightweight bookkeeping for the mmap'd ring area returned
//! by the kernel. Implementing fully-correct ring manipulation requires matching
//! the kernel's `xsk_ring_prod`/`xsk_ring_cons` layouts; here we provide a
//! structured place to implement those wrappers and a few safe placeholders so
//! higher-level code can be integrated incrementally.

use std::ptr::NonNull;

#[repr(C)]
#[derive(Clone, Copy, Debug)]
pub struct XskMmapOffsets {
    pub rx: u64,
    pub rx_desc: u64,
    pub tx: u64,
    pub tx_desc: u64,
    pub fill: u64,
    pub fill_desc: u64,
    pub comp: u64,
    pub comp_desc: u64,
}

pub struct RingMmap {
    base: NonNull<u8>,
    size: usize,
    offsets: XskMmapOffsets,
}

impl RingMmap {
    /// Construct a RingMmap wrapper around an mmap'ed pointer and reported offsets.
    pub unsafe fn new(map_ptr: *mut libc::c_void, map_size: usize, offs: XskMmapOffsets) -> Self {
        RingMmap { base: NonNull::new_unchecked(map_ptr as *mut u8), size: map_size, offsets: offs }
    }

    /// Access the raw base pointer
    pub fn base_ptr(&self) -> *mut u8 { self.base.as_ptr() }

    /// Report the mmap offsets
    pub fn offsets(&self) -> XskMmapOffsets { self.offsets }

    /// Placeholder: return how many RX descriptors appear available
    pub fn rx_available(&self) -> usize {
        // TODO: parse the RX ring consumer/producer indices from the mapped region
        0
    }

    /// Placeholder: pop up to `max` RX frame descriptors and return their offsets
    pub fn rx_pop(&self, _max: usize) -> Vec<u64> {
        // TODO: implement descriptor reads from ring memory
        Vec::new()
    }

    /// Placeholder: push `addrs` into the TX ring for transmission
    pub fn tx_push(&self, _addrs: &[u64]) -> usize {
        // TODO: implement descriptor writes to TX ring
        0
    }

    /// Low-level read of a u32 value at an mmap offset (little-endian).
    pub unsafe fn read_u32_at(&self, off: u64) -> u32 {
        let p = self.base.as_ptr().add(off as usize) as *const u32;
        u32::from_le(std::ptr::read_unaligned(p))
    }

    /// Low-level read of a u64 value at an mmap offset (little-endian).
    pub unsafe fn read_u64_at(&self, off: u64) -> u64 {
        let p = self.base.as_ptr().add(off as usize) as *const u64;
        u64::from_le(std::ptr::read_unaligned(p))
    }

    /// Low-level write of a u32 value at an mmap offset (little-endian).
    pub unsafe fn write_u32_at(&mut self, off: u64, v: u32) {
        let p = self.base.as_ptr().add(off as usize) as *mut u32;
        std::ptr::write_unaligned(p, v.to_le());
    }

    /// Return a borrow of the mapped slice at offset/len.
    /// Safety: caller must ensure the requested range is valid within the mmap.
    pub unsafe fn slice_at(&self, off: u64, len: usize) -> &'static [u8] {
        std::slice::from_raw_parts(self.base.as_ptr().add(off as usize), len)
    }
}
