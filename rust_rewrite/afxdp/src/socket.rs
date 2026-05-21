use std::sync::{Arc, Mutex};
use thiserror::Error;

use datapath::socket::XdpSocket as DatapathXdpSocket;

#[derive(Error, Debug)]
pub enum AfXdpError {
    #[error("initialization error: {0}")]
    Init(String),
}

/// Re-export a boxed trait object type compatible with the datapath crate's socket
pub type AfXdpSocket = Box<dyn DatapathXdpSocket + Send>;

/// A simple in-process mock socket useful for tests and CI.
pub struct MockSocket {
    frames: Arc<Mutex<Vec<Vec<u8>>>>,
    sent: Arc<Mutex<Vec<Vec<u8>>>>,
}

impl MockSocket {
    pub fn new(frames: Vec<Vec<u8>>) -> Self {
        Self { frames: Arc::new(Mutex::new(frames)), sent: Arc::new(Mutex::new(Vec::new())) }
    }

    pub fn take_sent(&self) -> Vec<Vec<u8>> { std::mem::take(&mut self.sent.lock().unwrap()) }
}

impl DatapathXdpSocket for MockSocket {
    fn poll(&mut self, _max: usize) -> Vec<Vec<u8>> { std::mem::take(&mut self.frames.lock().unwrap()) }
    fn send(&mut self, pkts: Vec<Vec<u8>>) -> Result<(), ()> { *self.sent.lock().unwrap() = pkts; Ok(()) }
}

// Provide a constructor that returns a boxed `datapath::socket::XdpSocket` object.
pub fn new_mock_socket(frames: Vec<Vec<u8>>) -> AfXdpSocket {
    Box::new(MockSocket::new(frames))
}

// --- Real socket skeleton --------------------------------------------------
// When built with `--features real` this module can be expanded to perform
// genuine AF_XDP UMEM allocation, ring setup and socket handling. For now we
// provide a thin wrapper type that can be completed later.

#[cfg(feature = "real")]
mod real {
    use super::*;
    use crate::umem::Umem;
    use std::os::unix::io::RawFd;

    pub struct RealSocket {
        ifname: String,
        queue_id: u32,
        fd: RawFd,
        _umem: Umem,
    }

    impl RealSocket {
        pub fn new(ifname: &str, queue_id: u32, umem_frame_size: usize, umem_pages: usize) -> Result<Self, AfXdpError> {
            // Resolve interface index and create AF_XDP socket, bind to queue, etc.
            // Implementing a fully working AF_XDP stack requires careful libc calls
            // and ring setup — this constructor prepares UMEM and opens a placeholder
            // socket FD. The detailed setup is left as the next step.
            let umem = Umem::new(umem_frame_size * umem_pages, umem_frame_size)
                .map_err(|e| AfXdpError::Init(format!("umem alloc: {}", e)))?;
            // create a raw socket placeholder (AF_XDP sock creation will go here)
            let fd = unsafe { libc::socket(libc::AF_XDP, libc::SOCK_RAW, 0) };
            if fd < 0 {
                return Err(AfXdpError::Init(std::io::Error::last_os_error().to_string()));
            }
            Ok(RealSocket { ifname: ifname.to_string(), queue_id, fd, _umem: umem })
        }
    }

    impl Drop for RealSocket {
        fn drop(&mut self) {
            unsafe { libc::close(self.fd); }
        }
    }

    impl datapath::socket::XdpSocket for RealSocket {
        fn poll(&mut self, _max: usize) -> Vec<Vec<u8>> {
            // TODO: implement RX ring consumption and produce owned packet buffers.
            Vec::new()
        }
        fn send(&mut self, _pkts: Vec<Vec<u8>>) -> Result<(), ()> {
            // TODO: implement TX ring enqueue and submission.
            Err(())
        }
    }
}

#[cfg(feature = "real")]
pub use real::RealSocket;
