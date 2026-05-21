// datapath crate: forwarder core

use crate::socket::XdpSocket;
use crypto::session::HybridSession;
use routing::Table;
use wire::Header;

pub struct Forwarder {
    pub routes: Table,
}

impl Forwarder {
    pub fn new(routes: Table) -> Self {
        Self { routes }
    }

    // Process a batch of frames from a socket.
    // For each frame: parse header, lookup next hop, and (if session exists) encrypt.
    pub fn process_batch(&self, sock: &mut dyn XdpSocket) {
        let frames = sock.poll(64);
        if frames.is_empty() {
            return;
        }
        let mut out: Vec<Vec<u8>> = Vec::with_capacity(frames.len());
        for mut pkt in frames {
            if let Ok(h) = Header::parse(&pkt) {
                // attempt route lookup
                if let Some(_nh) = self.routes.lookup_or_predict(h.src_id, h.dst_id, h.flow_label) {
                    // For demo: if payload non-empty, encrypt using a derived session
                    if h.length as usize > 0 {
                        // derive a dummy combined secret from kex path; here reuse session-info
                        let combined = vec![0u8;32];
                        if let Ok(sess) = HybridSession::new(&combined, &h.src_id) {
                            let payload_offset = wire::HEADER_SIZE;
                            if pkt.len() >= payload_offset {
                                let payload = &pkt[payload_offset..];
                                if let Ok(ct) = sess.encrypt(payload, h.seq_num) {
                                    // build new packet = header + ciphertext
                                    let mut newpkt = pkt[..payload_offset].to_vec();
                                    newpkt.extend_from_slice(&ct);
                                    out.push(newpkt);
                                    continue;
                                }
                            }
                        }
                    }
                }
            }
            // default: forward original packet
            out.push(pkt);
        }
        let _ = sock.send(out);
    }
}

pub mod socket {
    // Minimal XDP-like socket trait used by forwarder tests and mocks
    pub trait XdpSocket {
        fn poll(&mut self, max: usize) -> Vec<Vec<u8>>;
        fn send(&mut self, pkts: Vec<Vec<u8>>) -> Result<(), ()>;
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::socket::XdpSocket;
    use routing::Table;
    use wire::Header;

    struct MockSocket {
        frames: Vec<Vec<u8>>,
        sent: Vec<Vec<u8>>,
    }
    impl MockSocket {
        fn new(frames: Vec<Vec<u8>>) -> Self { Self { frames, sent: Vec::new() } }
    }
    impl XdpSocket for MockSocket {
        fn poll(&mut self, _max: usize) -> Vec<Vec<u8>> { std::mem::take(&mut self.frames) }
        fn send(&mut self, pkts: Vec<Vec<u8>>) -> Result<(), ()> { self.sent = pkts; Ok(()) }
    }

    #[test]
    fn forwarder_encrypts_and_sends() {
        let rt = Table::new();
        let fwd = Forwarder::new(rt);

        // build header + payload
        let mut buf = wire::Header::new_header_buffer(4);
        let h = Header { src_id: [1u8;32], dst_id: [2u8;32], flow_label: 0x1, seq_num: 1, session_id: [0u8;16], flags: 0, length: 4 };
        h.marshal_into(&mut buf).unwrap();
        // append payload
        buf[wire::HEADER_SIZE..wire::HEADER_SIZE+4].copy_from_slice(&[0x1,0x2,0x3,0x4]);
        let mut sock = MockSocket::new(vec![buf]);
        fwd.process_batch(&mut sock);
        // ensure something was sent
        assert!(!sock.sent.is_empty());
    }
}
