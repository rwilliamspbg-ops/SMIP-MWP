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

/-- Worker queues are assigned by adding the worker index to the base queue. -/
def queueForWorker (baseQueue worker : Nat) : Nat := baseQueue + worker

/-- Worker labels are rendered deterministically from the worker index. -/
def workerLabel (worker : Nat) : String := s!"worker-{worker}"

/-- A minimal buffer pool model with acquire/release round-trips. -/
abbrev BufferPool := List Nat

/-- Acquire the first available buffer from the pool. -/
def acquireBuffer : BufferPool → Option (Nat × BufferPool)
  | [] => none
  | b :: rest => some (b, rest)

/-- Return a buffer to the front of the pool. -/
def releaseBuffer (buffer : Nat) (pool : BufferPool) : BufferPool :=
  buffer :: pool

/-- The shard index is always within bounds. -/
theorem shardIndex_bounded (sid : Nat) : shardIndex sid < numSessionShards := by
  unfold shardIndex
  exact Nat.mod_lt _ (by decide)

/-- The base worker queue maps to itself. -/
theorem queueForWorker_zero (baseQueue : Nat) : queueForWorker baseQueue 0 = baseQueue := by
  simp [queueForWorker]

/-- Advancing to the next worker increments the assigned queue by one. -/
theorem queueForWorker_succ (baseQueue worker : Nat) :
    queueForWorker baseQueue (worker + 1) = queueForWorker baseQueue worker + 1 := by
  simp [queueForWorker, Nat.add_assoc, Nat.add_comm, Nat.add_left_comm]

/-- Worker labels are stable and deterministic for a given worker index. -/
theorem workerLabel_stable (worker : Nat) : workerLabel worker = s!"worker-{worker}" := by
  rfl

/-- Releasing a buffer and then acquiring it again returns the same buffer/pool pair. -/
theorem acquire_release_roundtrip (buffer : Nat) (pool : BufferPool) :
    acquireBuffer (releaseBuffer buffer pool) = some (buffer, pool) := by
  simp [acquireBuffer, releaseBuffer]

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