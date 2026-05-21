//! afxdp crate: AF_XDP UMEM and socket abstractions
//!
//! Provides a `MockSocket` for CI and a `RealSocket` skeleton that uses low-level
//! `libc`/`nix` primitives when built with `--features real`.

pub mod umem;
pub mod socket;
pub mod rings;

pub use socket::{AfXdpSocket, MockSocket};
// afxdp crate: XDP integration and mocks (scaffold)

pub fn available() -> bool {
    false
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn smoke() {
        assert!(!available());
    }
}
