import Lake

/-
Lightweight Lean spec for Crypto handshake/session invariants.
This file maps byte-sequence inputs to abstract session material and
proves basic deterministic and bounds properties. It's intentionally
lightweight: the concrete HKDF expansion is represented by `take`/slice
operations so properties can be discharged without crypto primitives.
 -/

namespace Smip

abbrev ByteSeq := List Nat

/-- Modeled session material: key (32), nonceBase (12), seqMask (64-bit) -/
structure SessionMaterial where
  key : ByteSeq
  nonceBase : ByteSeq
  seqMask : Nat

/-- Model: given combinedSecret and sessionInfo, produce session material.
    This mirrors `NewHybridSession` deterministic HKDF expansion in shape.
    For now we model expansion as deterministic slicing so we can prove
    determinism and simple length/bound properties. -/
def deriveSessionMaterial (combined sessionInfo : ByteSeq) : SessionMaterial :=
  { key := combined.take 32, nonceBase := sessionInfo.take 12, seqMask := 0 }

/- Lemmas about the shape and determinism of `deriveSessionMaterial` -/

theorem key_length_le (combined sessionInfo : ByteSeq) :
    (deriveSessionMaterial combined sessionInfo).key.length ≤ 32 := by
  simp [deriveSessionMaterial]
  apply Nat.min_le_left

theorem nonceBase_length_le (combined sessionInfo : ByteSeq) :
    (deriveSessionMaterial combined sessionInfo).nonceBase.length ≤ 12 := by
  simp [deriveSessionMaterial]
  apply Nat.min_le_left

theorem seqMask_zero (combined sessionInfo : ByteSeq) :
    (deriveSessionMaterial combined sessionInfo).seqMask = 0 := by
  simp [deriveSessionMaterial]

theorem derive_deterministic {c1 c2 s1 s2 : ByteSeq} (hc : c1 = c2) (hs : s1 = s2) :
    deriveSessionMaterial c1 s1 = deriveSessionMaterial c2 s2 := by
  simp [deriveSessionMaterial, hc, hs]

end Smip
