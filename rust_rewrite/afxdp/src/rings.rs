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
        unsafe {
            let offs = self.offsets;
            let rx_meta_off = offs.rx;
            let rx_desc_off = offs.rx_desc;

            // read producer and consumer indices
            let prod = self.read_u32_at(rx_meta_off) as u32;
            let cons = self.read_u32_at(rx_meta_off + 4) as u32;
            let avail = prod.wrapping_sub(cons) as usize;
            if avail == 0 { return Vec::new(); }

            // derive capacity from descriptor region length (tx_desc - rx_desc)
            let desc_region_bytes = (offs.tx_desc as i128 - offs.rx_desc as i128) as usize;
            let capacity = desc_region_bytes / std::mem::size_of::<u64>();
            let mask = capacity.saturating_sub(1);

            let to_take = std::cmp::min(avail, _max);
            let mut out = Vec::with_capacity(to_take);
            for i in 0..to_take {
                let idx = ((cons as usize + i) & mask) as usize;
                let d_off = rx_desc_off + (idx * std::mem::size_of::<u64>()) as u64;
                let desc = self.read_u64_at(d_off);
                out.push(desc);
            }

            // advance consumer index
            let new_cons = cons.wrapping_add(to_take as u32);
            self.write_u32_at(rx_meta_off + 4, new_cons);

            out
        }
    }

    /// Placeholder: push `addrs` into the TX ring for transmission
    pub fn tx_push(&self, _addrs: &[u64]) -> usize {
        unsafe {
            let offs = self.offsets;
            let tx_meta_off = offs.tx;
            let tx_desc_off = offs.tx_desc;

            let prod = self.read_u32_at(tx_meta_off) as u32;
            let cons = self.read_u32_at(tx_meta_off + 4) as u32;

            // compute capacity from descriptor region length (fill_desc - tx_desc)
            let desc_region_bytes = (offs.fill_desc as i128 - offs.tx_desc as i128) as usize;
            let capacity = desc_region_bytes / std::mem::size_of::<u64>();
            let mask = capacity.saturating_sub(1);

            let used = prod.wrapping_sub(cons) as usize;
            let free = capacity.saturating_sub(used);
            if free == 0 { return 0; }

            let to_push = std::cmp::min(free, _addrs.len());
            for i in 0..to_push {
                let idx = ((prod as usize + i) & mask) as usize;
                let d_off = tx_desc_off + (idx * std::mem::size_of::<u64>()) as u64;
                // write address
                let mut_self = &mut *(self as *const _ as *mut Self);
                mut_self.write_u64_at(d_off, _addrs[i]);
            }

            // advance producer
            let new_prod = prod.wrapping_add(to_push as u32);
            let mut_self = &mut *(self as *const _ as *mut Self);
            mut_self.write_u32_at(tx_meta_off, new_prod);
            to_push
        }
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

    /// Low-level write of a u64 value at an mmap offset (little-endian).
    pub unsafe fn write_u64_at(&mut self, off: u64, v: u64) {
        let p = self.base.as_ptr().add(off as usize) as *mut u64;
        std::ptr::write_unaligned(p, v.to_le());
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
