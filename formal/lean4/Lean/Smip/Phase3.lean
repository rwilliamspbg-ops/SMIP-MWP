namespace Smip

/-!
Phase 3 formalization skeleton.

This file contains initial datatypes and theorem stubs for routing properties.
TODO: replace the simplistic model with a verified network model (nodes, edges,
routing tables, and forwarding semantics).
-/

/- Richer network model for initial formalization. -/

/- Node is represented as a natural number id. -/
abbrev Node := Nat

/- Packet carries current location `loc`, destination `dst`, and a TTL counter. -/
structure Packet where
  loc : Node
  dst : Node
  ttl : Nat
  deriving Repr

/- A routing table maps (node, destination) -> optional next-hop node. -/
abbrev RoutingTable := Node → Node → Option Node

/- A network step forwards a packet according to the routing table if TTL > 0
   and a next-hop exists. Otherwise the packet is unchanged. -/
def forward_step (rt : RoutingTable) (p : Packet) : Packet :=
  match p.ttl with
  | 0 => p
  | Nat.succ k =>
    match rt p.loc p.dst with
    | none => { p with ttl := k }
    | some next => { loc := next, dst := p.dst, ttl := k }

/- Repeated forwarding: apply `n` steps. -/
def forward_n (rt : RoutingTable) : Nat → Packet → Packet
  | 0, p => p
  | Nat.succ n, p => forward_n rt n (forward_step rt p)

/- Lemma: Each forward_step decreases TTL by exactly 1 when TTL > 0. -/
theorem ttl_decreases_when_forwarded (rt : RoutingTable) (p : Packet) :
  p.ttl > 0 → (forward_step rt p).ttl = p.ttl - 1 := by
  intro h
  cases p.ttl
  · contradiction
  case succ k =>
    simp [forward_step]
    -- both branches set ttl := k (which equals p.ttl - 1)
    cases rt p.loc p.dst <;> simp

/- Corollary: after `p.ttl` steps, TTL is zero. -/
theorem ttl_reaches_zero (rt : RoutingTable) (p : Packet) :
  (forward_n rt p.ttl p).ttl = 0 := by
  induction p.ttl with
  | zero => simp [forward_n]
  | succ k ih =>
    simp [forward_n]
    have : (forward_step rt p).ttl = p.ttl - 1 := ttl_decreases_when_forwarded rt p (by simp)
    simp [this]
    apply ih

/- Definition (loop-freedom relative to TTL): there is no infinite forwarding
   sequence because TTL is strictly decreasing and bounded below by 0. -/
theorem no_infinite_forwarding (rt : RoutingTable) (p : Packet) :
  ∃ n, (forward_n rt n p).ttl = 0 := by
  use p.ttl
  exact ttl_reaches_zero rt p

/- Placeholder for a stronger property: routing loop-freedom independent of TTL.
   That requires proving absence of cycles in the routing graph or that routing
   strictly progresses along an acyclic metric. -/
theorem routing_loop_free_stronger : True := by trivial

end Smip
