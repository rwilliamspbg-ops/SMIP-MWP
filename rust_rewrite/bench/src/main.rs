use std::time::{Duration, Instant};
use bench::alloc_and_fill;

fn run_fixed_iters(size: usize, iters: usize) -> Duration {
    let start = Instant::now();
    for _ in 0..iters {
        let _ = alloc_and_fill(size);
    }
    start.elapsed()
}

fn main() {
    // Deterministic micro-benchmarks: fixed iteration counts give reproducible timing baselines
    let sizes = [1024usize, 8 * 1024, 64 * 1024];
    let iters = 100usize; // small, deterministic
    println!("Bench runner: {} iterations per size", iters);
    for &s in &sizes {
        let elapsed = run_fixed_iters(s, iters);
        let micros = elapsed.as_micros();
        let avg_us = micros as f64 / iters as f64;
        let bytes_per_sec = (s as f64) / (avg_us / 1_000_000.0);
        println!("size={} avg_us={:.2} bytes/sec={:.2}", s, avg_us, bytes_per_sec);
    }
}
