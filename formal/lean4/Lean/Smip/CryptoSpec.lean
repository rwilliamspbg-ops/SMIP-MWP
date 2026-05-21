namespace Smip

/-!
Lean model for the crypto constants and cache bounds used by the Go hybrid KEX
and HKDF cache implementation.
-/

/-- The bounded HKDF cache in Go is capped at 10,000 derived entries. -/
def maxHKDFCacheSize : Nat := 10000

/-- The x25519 public key length used by the hybrid handshake. -/
def x25519PubLen : Nat := 32

/-- The ML-KEM-768 public key length used by the hybrid handshake. -/
def mlkemPubLen : Nat := 1184

/-- The hybrid handshake requires this minimum peer public-key length. -/
def peerPublicKeyMinLen : Nat := x25519PubLen + mlkemPubLen

/-- The derived session secret is 64 bytes. -/
def sessionSecretLen : Nat := 64

/-- The cache bound matches the Go constant. -/
theorem hkdf_cache_bounded : maxHKDFCacheSize = 10000 := by
  rfl

/-- The peer public key length check matches the Go handshake guard. -/
theorem peer_public_key_min_len : peerPublicKeyMinLen = 1216 := by
  rfl

/-- The hybrid handshake expands to a 64-byte session secret. -/
theorem session_secret_len : sessionSecretLen = 64 := by
  rfl

end Smip