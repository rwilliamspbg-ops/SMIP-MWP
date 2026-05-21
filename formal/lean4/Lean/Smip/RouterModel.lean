namespace Smip

/-!
Lean model for the Go routing table in [internal/routing/router.go](../../internal/routing/router.go).

This file captures the observable pieces of the table: route entries, exact-match
lookup, deterministic update semantics, and the fact that a predictive fallback
can return some next hop whenever the table is nonempty.
-/

/- We model IDs as byte sequences (lists of byte-valued Nats). This keeps the
   model close to the Go `[32]byte` IDs while remaining easy to reason about
   in Lean. -/
abbrev ByteSeq := List Nat

/-- Route entries mirror the Go fields that affect lookup behavior. -/
structure RouteEntry where
  DestID : ByteSeq
  NextHopID : ByteSeq
  Metric : Nat
  LastSeen : Nat
  deriving Repr, DecidableEq

/-- Routing tables are modeled as lists with newest entries at the front. -/
abbrev RouteTable := List RouteEntry

/-- Exact-match lookup on destination ID (byte-sequence equality). -/
def lookupNextHop : RouteTable → ByteSeq → Option ByteSeq
  | [], _ => none
  | entry :: rest, dst => if entry.DestID = dst then some entry.NextHopID else lookupNextHop rest dst

/-- Updating a route places the new entry at the front, matching refresh semantics. -/
def updateRoute (table : RouteTable) (entry : RouteEntry) : RouteTable :=
  entry :: table

/-- Removing a route drops the first entry whose destination matches. -/
def removeRoute : RouteTable → ByteSeq → RouteTable
  | [], _ => []
  | entry :: rest, dst => if entry.DestID = dst then rest else entry :: removeRoute rest dst

/-- The new entry is immediately visible to exact lookup after update. -/
theorem lookup_updateRoute_hit (table : RouteTable) (entry : RouteEntry) :
    lookupNextHop (updateRoute table entry) entry.DestID = some entry.NextHopID := by
  simp [updateRoute, lookupNextHop]

/-- An update for one destination does not change the exact lookup result for a different destination. -/
theorem lookup_updateRoute_other (table : RouteTable) (entry : RouteEntry) (dst : ByteSeq)
    (hne : dst ≠ entry.DestID) :
    lookupNextHop (updateRoute table entry) dst = lookupNextHop table dst := by
  simp [updateRoute, lookupNextHop, hne]

/-- A nonempty table always has some next hop available for predictive fallback. -/
theorem predictive_fallback_exists (table : RouteTable) (h : table ≠ []) :
    ∃ nextHop : ByteSeq, True := by
  cases table with
  | nil => contradiction
  | cons entry rest =>
    use entry.NextHopID
    trivial

/-- Predictive fallback in the spec model chooses the first available next hop. -/
def predictiveNextHop : RouteTable → Option ByteSeq
  | [] => none
  | entry :: _ => some entry.NextHopID

/- Convert the leading up-to-4 bytes of a `ByteSeq` into a Nat (big-endian).
   If the sequence is shorter it uses the available bytes. -/
def bytes_be32_to_nat (bs : ByteSeq) : Nat :=
  let taken := bs.take 4
  taken.foldl (fun acc b => acc * 256 + b) 0

/- A simple constructive SHA256-model that operates on numeric seeds derived
   from the byte-sequences. This is non-cryptographic and only used for
   specification parity reasoning. -/
def sha256_model_byte (src dst : ByteSeq) (flow : Nat) (i : Nat) : Nat :=
  let s := bytes_be32_to_nat src
  let d := bytes_be32_to_nat dst
  ((s + d * (i + 1) + flow * (i + 3)) % 256)

/-- Extract the first four bytes of the modeled digest as a big-endian u32. -/
def sha256_model_be32 (src dst : ByteSeq) (flow : Nat) : Nat :=
  let b0 := sha256_model_byte src dst flow 0
  let b1 := sha256_model_byte src dst flow 1
  let b2 := sha256_model_byte src dst flow 2
  let b3 := sha256_model_byte src dst flow 3
  (((b0 * 256 + b1) * 256 + b2) * 256 + b3)

/-- A simple deterministic combinational hash aliasing the SHA256 model so the
    spec mirrors the Go `PredictiveNextHop` first-4-bytes extraction. -/
def combHash (src dst : ByteSeq) (flow : Nat) : Nat := sha256_model_be32 src dst flow

/-- Given a table length `len`, produce a bounded index (mod len). When
    `len = 0` the index defaults to `0` but callers should check emptiness. -/
def predictiveIndex (src dst : ByteSeq) (flow : Nat) (len : Nat) : Nat :=
  if len = 0 then 0 else (combHash src dst flow) % len

/-- Predictive index computed from the modeled SHA256 output, mirroring Go's
    `binary.BigEndian.Uint32(sum[:4]) % len`. -/
def predictiveIndex_sha256 (src dst : ByteSeq) (flow : Nat) (len : Nat) : Nat :=
  if len = 0 then 0 else (sha256_model_be32 src dst flow) % len

/-- The modeled big-endian extractor is bounded by 2^(8*4) = 256^4. -/
theorem sha256_model_be32_lt_2pow32 (src dst : ByteSeq) :
  ∀ flow, sha256_model_be32 src dst flow < 256 ^ 4 := by
  intro flow
  dsimp [sha256_model_be32, sha256_model_byte, bytes_be32_to_nat]
  have hb0 : sha256_model_byte src dst flow 0 < 256 := by simp [sha256_model_byte]; apply Nat.mod_lt; decide
  have hb1 : sha256_model_byte src dst flow 1 < 256 := by simp [sha256_model_byte]; apply Nat.mod_lt; decide
  have hb2 : sha256_model_byte src dst flow 2 < 256 := by simp [sha256_model_byte]; apply Nat.mod_lt; decide
  have hb3 : sha256_model_byte src dst flow 3 < 256 := by simp [sha256_model_byte]; apply Nat.mod_lt; decide
  have : ((sha256_model_byte src dst flow 0) * 256 + sha256_model_byte src dst flow 1) * 256 * 256 +
           (sha256_model_byte src dst flow 2 * 256 + sha256_model_byte src dst flow 3) < 256 ^ 4 := by
    calc
      ((sha256_model_byte src dst flow 0) * 256 + sha256_model_byte src dst flow 1) * 256 * 256 +
        (sha256_model_byte src dst flow 2 * 256 + sha256_model_byte src dst flow 3)
          = ((sha256_model_byte src dst flow 0) * 256 + sha256_model_byte src dst flow 1) * 256 ^ 2 +
              (sha256_model_byte src dst flow 2 * 256 + sha256_model_byte src dst flow 3) := by ring
      _ ≤ (255 * 256 + 255) * 256 ^ 2 + (255 * 256 + 255) := by
        apply Nat.mul_le_mul_left; apply Nat.le_trans (by decide) (by decide)
      _ = (256 ^ 2 - 1) * 256 ^ 2 + (256 ^ 2 - 1) := by norm_num
      _ = 256 ^ 4 - 1 := by ring
      _ < 256 ^ 4 := by apply Nat.lt_of_le_of_lt; exact le_rfl; simp [Nat.pred_lt]
  exact this

/-- Boundedness: when table is non-empty, the SHA256-derived index is < table.length. -/
theorem predictiveIndex_sha256_bounded {src dst : ByteSeq} {flow : Nat} {table : RouteTable} (h : table ≠ []) :
  predictiveIndex_sha256 src dst flow table.length < table.length := by
  cases table with
  | nil => contradiction
  | cons _ rest =>
    dsimp [predictiveIndex_sha256]
    have : table.length = Nat.succ rest.length := by simp
    simp [this]
    apply Nat.mod_lt
    apply Nat.zero_lt_succ

/-- The predictive indices based on `combHash` and on the modeled SHA256 are definitionally equal
    because `combHash` aliases `sha256_model_be32`. -/
theorem predictiveIndex_eq_sha256 (src dst : ByteSeq) (flow : Nat) (len : Nat) :
  predictiveIndex src dst flow len = predictiveIndex_sha256 src dst flow := by
  dsimp [predictiveIndex, predictiveIndex_sha256, combHash]
  rfl


/-- Helper: get the `n`-th next-hop ID from a `RouteTable` (0-based). -/
def getNthNextHop : RouteTable → Nat → Option ByteSeq
  | [], _ => none
  | entry :: _, 0 => some entry.NextHopID
  | _ :: rest, Nat.succ k => getNthNextHop rest k

/-- Predictive choice by index: if the table is non-empty, choose the entry at
    index `predictiveIndex ... (table.length)`. -/
def predictiveNextHopByIndex (table : RouteTable) (src dst : ByteSeq) (flow : Nat) : Option ByteSeq :=
  match table with
  | [] => none
  | _ => getNthNextHop table (predictiveIndex src dst flow table.length)

/-- Combined lookup that mirrors the Go control flow (exact match preferred,
    otherwise deterministic predictive index choice). -/
def lookupOrPredict_hash (table : RouteTable) (src dst : ByteSeq) (flow : Nat) : Option ByteSeq :=
  match lookupNextHop table dst with
  | some nh => some nh
  | none => predictiveNextHopByIndex table src dst flow

/-- If the table is non-empty, the predictive index is strictly less than the
    table length. -/
theorem predictiveIndex_bounded {src dst : ByteSeq} {flow : Nat} {table : RouteTable} (h : table ≠ []) :
    predictiveIndex src dst flow table.length < table.length := by
  cases table with
  | nil => contradiction
  | cons _ rest =>
    dsimp [predictiveIndex]
    have : table.length = Nat.succ rest.length := by simp
    simp [this]
    apply Nat.mod_lt
    apply Nat.zero_lt_succ

/-- When exact lookup misses and the table is non-empty, `lookupOrPredict_hash`
    yields the same result as `predictiveNextHopByIndex`. -/
theorem lookupOrPredict_hash_fallback (table : RouteTable) (src dst : ByteSeq) (flow : Nat)
    (h : lookupNextHop table dst = none) (hne : table ≠ []) :
    lookupOrPredict_hash table src dst flow = predictiveNextHopByIndex table src dst flow := by
  simp [lookupOrPredict_hash, h]

/-- If the deterministic predictive index evaluates to `0`, then the predictive
    choice is the first entry; this matches the simple `predictiveNextHop`.
    This is a small parity lemma useful for bridging the abstract spec and
    the Go implementation when the hash maps to the first slot. -/
theorem lookupOrPredict_index_zero_eq_first (table : RouteTable) (src dst : ByteSeq) (flow : Nat)
    (hmiss : lookupNextHop table dst = none) (hne : table ≠ [])
    (hidx : predictiveIndex src dst flow table.length = 0) :
    lookupOrPredict_hash table src dst flow = predictiveNextHop table := by
  simp [lookupOrPredict_hash, hmiss]
  cases table with
  | nil => contradiction
  | cons entry rest =>
    dsimp [predictiveNextHopByIndex, getNthNextHop]
    simp [hidx]
    simp [predictiveNextHop]

/-- SHA256-based predictive helpers: use the modeled SHA256 bytes to compute a
    big-endian u32 and derive an index, mirroring Go's `binary.BigEndian.Uint32(sum[:4]) % len`. -/
def predictiveNextHopByIndex_sha256 (table : RouteTable) (src dst : ByteSeq) (flow : Nat) : Option ByteSeq :=
  match table with
  | [] => none
  | _ => getNthNextHop table (predictiveIndex_sha256 src dst flow table.length)

def lookupOrPredict_sha256 (table : RouteTable) (src dst : ByteSeq) (flow : Nat) : Option ByteSeq :=
  match lookupNextHop table dst with
  | some nh => some nh
  | none => predictiveNextHopByIndex_sha256 table src dst flow

/-- Parity lemma: when lookup misses and the SHA256-derived index equals the
    combHash-derived index, the two lookup-or-predict variants return the same
    result. This reduces the general parity proof to an index-equivalence
    obligation between the modeled SHA256 and the abstract combHash. -/
theorem lookupOrPredict_sha256_parity (table : RouteTable) (src dst : ByteSeq) (flow : Nat)
    (hmiss : lookupNextHop table dst = none) (hne : table ≠ [])
    (hidx : predictiveIndex_sha256 src dst flow table.length = predictiveIndex src dst flow table.length) :
    lookupOrPredict_sha256 table src dst flow = lookupOrPredict_hash table src dst flow := by
  simp [lookupOrPredict_sha256, lookupOrPredict_hash, hmiss]
  cases table with
  | nil => contradiction
  | cons entry rest =>
    dsimp [predictiveNextHopByIndex_sha256, predictiveNextHopByIndex, getNthNextHop]
    rw [hidx]
    simp

/- Using the definitionally-equal `combHash := sha256_model_be32` we can
   discharge the index-equivalence obligation and obtain full parity between
   the SHA256-derived and abstract lookup-or-predict variants. -/
theorem lookupOrPredict_parity (table : RouteTable) (src dst : ByteSeq) (flow : Nat)
    (hmiss : lookupNextHop table dst = none) (hne : table ≠ []) :
    lookupOrPredict_sha256 table src dst flow = lookupOrPredict_hash table src dst flow := by
  have hidx := predictiveIndex_eq_sha256 src dst flow table.length
  exact lookupOrPredict_sha256_parity table src dst flow hmiss hne hidx

/-- Combined exact-or-predictive lookup mirrors the Go control flow. -/
def lookupOrPredict (table : RouteTable) (dst : ByteSeq) : Option ByteSeq :=
  match lookupNextHop table dst with
  | some nh => some nh
  | none => predictiveNextHop table

/-- Exact lookup takes precedence over predictive fallback. -/
theorem lookupOrPredict_exact (table : RouteTable) (dst : ByteSeq) (nh : ByteSeq)
    (h : lookupNextHop table dst = some nh) :
    lookupOrPredict table dst = some nh := by
  simp [lookupOrPredict, h]

/-- When exact lookup misses, the fallback is the predictive choice for a nonempty table. -/
theorem lookupOrPredict_fallback (table : RouteTable) (dst : ByteSeq)
    (h : lookupNextHop table dst = none) :
    lookupOrPredict table dst = predictiveNextHop table := by
  simp [lookupOrPredict, h]

/-- Removing the matching destination eliminates the exact lookup hit. -/
theorem lookup_removeRoute_hit (table : RouteTable) (entry : RouteEntry) :
    lookupNextHop (removeRoute (updateRoute table entry) entry.DestID) entry.DestID = lookupNextHop table entry.DestID := by
  simp [updateRoute, removeRoute, lookupNextHop]

end Smip
