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

/- Stronger routing loop-freedom when a numeric distance metric strictly
   decreases at each hop (well-formed routing table).

/- Distance metric: `dist n d` gives a natural-number distance from node `n`
   to destination `d`. -/
abbrev Distance := Node → Node → Nat

/- Routing well-formedness: if `rt n d = some next` then `dist next d < dist n d`.
   This enforces that following the routing table strictly decreases the
   numeric distance to the destination, ruling out cycles. -/
def routing_wf (rt : RoutingTable) (dist : Distance) : Prop :=
  ∀ n d next, rt n d = some next → dist next d < dist n d

/- If the routing table is well-formed, a single forwarding step (when a
   next-hop exists) strictly decreases the distance to the destination. -/
theorem dist_decreases_on_forward {rt : RoutingTable} {dist : Distance}
    (wf : routing_wf rt dist) (p : Packet) :
  match rt p.loc p.dst with
  | none => True
  | some next => dist next p.dst < dist p.loc p.dst := by
  cases rt p.loc p.dst <;> simp at *
  exact wf p.loc p.dst _ rfl

/- If `rt` is well-formed, then starting from any packet `p`, after at most
   `dist p.loc p.dst` forwarding steps you will reach a node where the routing
   table does not provide a next-hop (or you are at the destination). This
   implies there are no routing cycles under the well-formedness invariant. -/
theorem bounded_hops_to_none_or_dst {rt : RoutingTable} {dist : Distance}
    (wf : routing_wf rt dist) (p : Packet) :
  ∃ k, k ≤ dist p.loc p.dst ∧
    (let q := forward_n rt k p; rt q.loc q.dst = none ∨ q.loc = q.dst) := by
  -- strong induction on the distance measure
  apply Nat.strong_induction_on (dist p.loc p.dst)
  intro m ih
  -- generalize over the packet whose distance equals `m` (we'll only apply ih
  -- for strictly smaller measures)
  have : dist p.loc p.dst = m := by simp_all
  by_cases h : rt p.loc p.dst = none
  · use 0
    constructor
    · simp
    · simp [h]
  · -- there is a next hop
    cases rt p.loc p.dst with
    | none => contradiction
    | some next =>
      have dlt : dist next p.dst < dist p.loc p.dst := wf p.loc p.dst next rfl
      have lt : dist next p.dst < m := by
        -- dist p.loc p.dst = m, so this is equivalent
        simp [this] at dlt
        exact dlt
      -- construct packet at next
      let p' := { loc := next, dst := p.dst, ttl := p.ttl }
      -- apply IH for the smaller distance `dist next p.dst` (which is < m)
      have ih_app := ih (dist next p.dst) (by exact lt)
      -- ih_app is a function that given a packet with distance `dist next p.dst`
      -- returns the existential; apply it to `p'`.
      obtain ⟨k', hk', Hk'⟩ := ih_app p'
      use k' + 1
      constructor
      · apply Nat.succ_le_succ hk'
      · simp [forward_n]
        -- forward one step then k' steps
        have : forward_n rt (k' + 1) p = forward_n rt k' (forward_step rt p) := by simp [forward_n]
        simp [this]
        exact Hk'

/- Stronger loop-freedom placeholder (proves True until more properties are
   formalized). -/
theorem routing_loop_free_stronger : True := by trivial
end Smip
