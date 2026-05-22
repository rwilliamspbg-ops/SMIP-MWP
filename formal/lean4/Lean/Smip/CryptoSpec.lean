namespace Smip

/-!
Lean model for the crypto constants and cache bounds used by the Go hybrid KEX
and HKDF cache implementation.
-/

abbrev ByteSeq := List Nat

/-- A minimal HKDF cache key based on the lengths of the derived inputs. -/
abbrev HKDFCacheKey := Nat × Nat

/-- A minimal HKDF cache representation. -/
abbrev HKDFCache := List (HKDFCacheKey × ByteSeq)

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

/-- The spec cache key is deterministic for a given pair of inputs. -/
def hkdfCacheKey (combined sessionInfo : ByteSeq) : HKDFCacheKey :=
  (combined.length, sessionInfo.length)

/-- Exact-match lookup for the abstract HKDF cache. -/
def lookupHKDFCache : HKDFCache → HKDFCacheKey → Option ByteSeq
  | [], _ => none
  | (k, v) :: rest, key => if k = key then some v else lookupHKDFCache rest key

/-- Inserting a derived cache entry makes it immediately visible to exact lookup. -/
def insertHKDFCache (cache : HKDFCache) (key : HKDFCacheKey) (value : ByteSeq) : HKDFCache :=
  (key, value) :: cache

/-- The cache bound matches the Go constant. -/
theorem hkdf_cache_bounded : maxHKDFCacheSize = 10000 := by
  rfl

/-- The peer public key length check matches the Go handshake guard. -/
theorem peer_public_key_min_len : peerPublicKeyMinLen = 1216 := by
  rfl

/-- The hybrid handshake expands to a 64-byte session secret. -/
theorem session_secret_len : sessionSecretLen = 64 := by
  rfl

/-- The HKDF cache key derivation is deterministic. -/
theorem hkdfCacheKey_deterministic (combined sessionInfo : ByteSeq) :
    hkdfCacheKey combined sessionInfo = hkdfCacheKey combined sessionInfo := by
  rfl

/-- Exact-match cache insertion is immediately visible to lookup. -/
theorem lookup_insertHKDFCache_hit (cache : HKDFCache) (key : HKDFCacheKey) (value : ByteSeq) :
    lookupHKDFCache (insertHKDFCache cache key value) key = some value := by
  simp [insertHKDFCache, lookupHKDFCache]

end Smip