pub mod kex;
pub mod session;

#[cfg(test)]
mod tests {
    use super::kex::*;
    use super::session::*;

    #[test]
    fn kex_and_session_flow() {
        let k = HybridKEX::new().expect("kex new");
        let pubk = k.public_key();
        let combined = k.handshake(&pubk).expect("handshake");
        let sess = HybridSession::new(&combined, b"session-info").expect("session");
        let ct = sess.encrypt(b"hello", 1).expect("encrypt");
        let pt = sess.decrypt(&ct, 1).expect("decrypt");
        assert_eq!(pt, b"hello");
    }
}
