// bench crate: benchmarks and MoonGen harness adapters (scaffold)

/// Allocate a `Vec<u8>` of `size` bytes and fill it with a deterministic pattern.
pub fn alloc_and_fill(size: usize) -> Vec<u8> {
    let mut v = Vec::with_capacity(size);
    // Safety: set_len after reserve so we can write into the buffer directly
    unsafe { v.set_len(size); }
    for i in 0..size {
        v[i] = (i & 0xFF) as u8;
    }
    v
}

pub fn run_bench() {
    // placeholder for integration with other harnesses
}
