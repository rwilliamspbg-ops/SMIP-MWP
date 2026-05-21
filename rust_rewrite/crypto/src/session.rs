use aes_gcm::{Aes256Gcm, aead::{Aead, KeyInit, OsRng}};
use chacha20poly1305::ChaCha20Poly1305;
use hkdf::Hkdf;
use sha2::Sha256;
use parking_lot::RwLock;
use std::collections::HashMap;
use thiserror::Error;

pub const TAG_SIZE: usize = 16;
pub const KEY_SIZE: usize = 32;
pub const NONCE_SIZE: usize = 12;

#[derive(Error, Debug)]
pub enum SessionError {
    #[error("buffer too small")]
    BufferTooSmall,
    #[error("aead error")]
    AeadError,
}

#[derive(Clone)]
struct CacheEntry {
    key: [u8; KEY_SIZE],
    nonce_base: [u8; NONCE_SIZE],
    seq_mask: u64,
}

type CacheMap = RwLock<HashMap<[u8;32], CacheEntry>>;

fn derive_cache_key(combined_secret: &[u8], session_info: &[u8]) -> [u8;32] {
    use sha2::Digest;
    let mut h = Sha256::new();
    h.update(combined_secret);
    h.update(session_info);
    let sum = h.finalize();
    let mut out = [0u8;32];
    out.copy_from_slice(&sum);
    out
}

static HKDF_CACHE: once_cell::sync::Lazy<CacheMap> = once_cell::sync::Lazy::new(|| RwLock::new(HashMap::new()));

pub struct HybridSession {
    // prefer AES-GCM when available
    aead: Aes256Gcm,
    nonce_base: [u8; NONCE_SIZE],
    seq_mask: u64,
}

impl HybridSession {
    pub fn new(combined_secret: &[u8], session_info: &[u8]) -> Result<Self, SessionError> {
        let cache_key = derive_cache_key(combined_secret, session_info);
        {
            let map = HKDF_CACHE.read();
            if let Some(e) = map.get(&cache_key) {
                let aead = Aes256Gcm::new_from_slice(&e.key).map_err(|_| SessionError::AeadError)?;
                return Ok(HybridSession { aead, nonce_base: e.nonce_base, seq_mask: e.seq_mask });
            }
        }

        let hk = Hkdf::<Sha256>::new(None, combined_secret);
        let mut key = [0u8; KEY_SIZE];
        hk.expand(b"smip-mwp-session-v1", &mut key).map_err(|_| SessionError::AeadError)?;

        // derive nonce base and seq mask by expanding more bytes
        let mut nonce_base = [0u8; NONCE_SIZE];
        let mut mask_bytes = [0u8;8];
        let mut info = [0u8; 0];
        let hk2 = Hkdf::<Sha256>::new(None, &key);
        hk2.expand(b"nonce", &mut nonce_base).map_err(|_| SessionError::AeadError)?;
        hk2.expand(b"mask", &mut mask_bytes).map_err(|_| SessionError::AeadError)?;

        let seq_mask = u64::from_be_bytes(mask_bytes);

        let aead = Aes256Gcm::new_from_slice(&key).map_err(|_| SessionError::AeadError)?;

        let entry = CacheEntry { key, nonce_base, seq_mask };
        {
            let mut map = HKDF_CACHE.write();
            map.insert(cache_key, entry.clone());
        }

        Ok(HybridSession { aead, nonce_base, seq_mask })
    }

    fn build_nonce(&self, seq: u64) -> [u8; NONCE_SIZE] {
        let mut nonce = [0u8; NONCE_SIZE];
        nonce.copy_from_slice(&self.nonce_base);
        let mut last8 = u64::from_be_bytes(nonce[4..12].try_into().unwrap());
        last8 ^= seq ^ self.seq_mask;
        nonce[4..12].copy_from_slice(&last8.to_be_bytes());
        nonce
    }

    pub fn encrypt(&self, plaintext: &[u8], seq: u64) -> Result<Vec<u8>, SessionError> {
        let nonce = self.build_nonce(seq);
        let ct = self.aead.encrypt(nonce.as_ref().into(), plaintext).map_err(|_| SessionError::AeadError)?;
        Ok(ct)
    }

    pub fn decrypt(&self, ciphertext: &[u8], seq: u64) -> Result<Vec<u8>, SessionError> {
        let nonce = self.build_nonce(seq);
        let pt = self.aead.decrypt(nonce.as_ref().into(), ciphertext).map_err(|_| SessionError::AeadError)?;
        Ok(pt)
    }
}
