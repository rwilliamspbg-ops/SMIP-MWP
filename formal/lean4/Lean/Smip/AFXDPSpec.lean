namespace Smip

/-!
Lean model for the AF_XDP forwarder lifecycle and shard-selection behavior.

The Go implementation uses a fixed shard count for sessions and a running flag
for start/stop control. This spec captures those observable invariants.
-/

/-- The forwarder uses 16 session shards. -/
def numSessionShards : Nat := 16

/-- Session shards are selected by a bounded modulus. -/
def shardIndex (sid : Nat) : Nat := sid % numSessionShards

/-- The shard index is always within bounds. -/
theorem shardIndex_bounded (sid : Nat) : shardIndex sid < numSessionShards := by
  unfold shardIndex
  exact Nat.mod_lt _ (by decide)

/-- Minimal forwarder lifecycle state. -/
structure ForwarderState where
  running : Bool
  activeWorkers : Nat
  deriving Repr, DecidableEq

/-- Starting a forwarder marks it running. -/
def start (s : ForwarderState) : ForwarderState :=
  { s with running := true }

/-- Stopping a forwarder marks it not running and clears active workers in the spec model. -/
def stop (s : ForwarderState) : ForwarderState :=
  { running := false, activeWorkers := 0 }

/-- Stopping twice is the same as stopping once. -/
theorem stop_idempotent (s : ForwarderState) : stop (stop s) = stop s := by
  simp [stop]

/-- Starting after stopping yields a running state again. -/
theorem start_after_stop_running (s : ForwarderState) : (start (stop s)).running = true := by
  simp [start, stop]

end Smip