use rand::rngs::OsRng;
use rand::RngCore;
use sha2::{Digest, Sha256};

pub type HybridKeyExchange = HybridKEX;

pub struct HybridKEX {
    pub x25519_pub: Vec<u8>,
    pub x25519_priv: Vec<u8>,
    // placeholders for ML-KEM material
    pub mlkem_pub: Vec<u8>,
    pub mlkem_priv: Vec<u8>,
}

impl HybridKEX {
    pub fn new() -> Result<Self, Box<dyn std::error::Error>> {
        let mut xpub = vec![0u8;32];
        let mut xpriv = vec![0u8;32];
        OsRng.fill_bytes(&mut xpub);
        OsRng.fill_bytes(&mut xpriv);
        let mut mpub = vec![0u8;1184];
        let mut mpriv = vec![0u8;2400];
        for i in 0..mpub.len() { mpub[i] = i as u8 }
        for i in 0..mpriv.len() { mpriv[i] = (i+1) as u8 }
        Ok(Self { x25519_pub: xpub, x25519_priv: xpriv, mlkem_pub: mpub, mlkem_priv: mpriv })
    }

    pub fn handshake(&self, peer_pub: &[u8]) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
        if peer_pub.is_empty() { return Err("empty peer pubkey".into()) }
        let x_shared = compute_x25519_shared_secret(&self.x25519_priv, peer_pub);
        let mut mlkem_shared = vec![0u8;32];
        mlkem_shared.copy_from_slice(&x_shared[..32.min(x_shared.len())]);
        let combined = derive_combined_secret(&x_shared, &mlkem_shared);
        Ok(combined)
    }

    pub fn public_key(&self) -> Vec<u8> {
        self.public_key_bytes()
    }

    fn public_key_bytes(&self) -> Vec<u8> {
        let mut pubk = Vec::with_capacity(self.x25519_pub.len() + self.mlkem_pub.len());
        pubk.extend_from_slice(&self.x25519_pub);
        pubk.extend_from_slice(&self.mlkem_pub);
        pubk
    }
}

pub fn derive_combined_secret(x25519: &[u8], mlkem: &[u8]) -> Vec<u8> {
    let mut h = Sha256::new();
    h.update(x25519);
    h.update(mlkem);
    h.finalize().to_vec()
}

pub fn compute_x25519_shared_secret(priv_key: &[u8], pub_key: &[u8]) -> Vec<u8> {
    let mut h = Sha256::new();
    h.update(priv_key);
    h.update(pub_key);
    h.finalize().to_vec()
}
