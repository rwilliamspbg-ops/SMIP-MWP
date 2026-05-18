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

/- Stronger routing loop-freedom when a numeric distance metric strictly
  decreases at each hop (well-formed routing table).

/- Graph model: neighbors function returning adjacency list for a node. -/
abbrev Neighbors := Node → List Node

/- pick_next searches the neighbors list for a neighbor with strictly
  smaller distance to destination and returns the first such neighbor. -/
partial def pick_next? (ns : List Node) (n d : Node) (dist : Distance) : Option Node :=
  match ns with
  | [] => none
  | h::t => if dist h d < dist n d then some h else pick_next? t n d dist

/- Build a routing table from neighbors by picking a next-hop that reduces
  the distance if one exists. -/
def build_rt (neighbors : Neighbors) (dist : Distance) : RoutingTable :=
  fun n d => pick_next? (neighbors n) n d dist

/- If `build_rt` yields `some next`, then by construction `dist next d < dist n d`.
  This proves `routing_wf` for `build_rt`. -/
theorem build_rt_wf (neighbors : Neighbors) (dist : Distance) :
  routing_wf (build_rt neighbors dist) dist := by
  intro n d next h
  dsimp [build_rt] at h
  -- unfold pick_next? via cases on `neighbors n`
  have hn := neighbors n
  generalize hn_val : neighbors n = hn
  revert hn_val h
  intro ns hn_val h
  induction ns with
  | nil => simp [pick_next?] at h; contradiction
  | cons hd tl ih =>
   simp [pick_next?] at h
   by_cases c : dist hd d < dist n d
   · simp [c] at h; injection h with h1; subst h1; exact c
   · simp [c] at h; exact ih rfl h

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

/- The destination field is preserved by forwarding steps and by repeated
   forwarding. -/
theorem dst_preserved_by_forward_step (rt : RoutingTable) (p : Packet) :
  (forward_step rt p).dst = p.dst := by
  simp [forward_step]

theorem dst_preserved_by_forward_n (rt : RoutingTable) :
  ∀ n p, (forward_n rt n p).dst = p.dst := by
  intro n
  induction n with
  | zero => intro p; simp [forward_n]
  | succ k ih =>
    intro p
    simp [forward_n]
    apply ih

/- Distance decreases for each concrete forwarding step when a next-hop
   exists; extend to repeated steps showing a strict chain of decreases. -/
theorem forward_step_decreases_dist {rt : RoutingTable} {dist : Distance}
    (wf : routing_wf rt dist) (p : Packet) :
  match rt p.loc p.dst with
  | none => True
  | some next => dist (forward_step rt p).loc p.dst < dist p.loc p.dst := by
  cases rt p.loc p.dst <;> simp at *
  -- when `some next` the forward_step loc becomes `next`
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

      /- Stronger loop-freedom placeholder removed; see top-level theorem.
        (A proper statement is added at top-level after this proof.) -/

      /- Concrete example: a small acyclic graph and its distance-to-destination.
         Nodes: 0,1,2,3. Edges: 0->1,0->2,1->3,2->3. Destination node is 3. -/
      def example_neighbors : Neighbors := fun n =>
        match n with
        | 0 => [1,2]
        | 1 => [3]
        | 2 => [3]
        | 3 => []
        | _ => []

      def example_dist : Distance := fun n d =>
        if d = 3 then
          match n with
          | 0 => 2
          | 1 => 1
          | 2 => 1
          | 3 => 0
          | _ => 1000
        else 1000

      def example_rt : RoutingTable := build_rt example_neighbors example_dist

      theorem example_build_rt_wf : routing_wf example_rt example_dist :=
        build_rt_wf example_neighbors example_dist

      theorem example_reaches_dst (p : Packet) (h : p.dst = 3) (hloc : p.loc < 4) :
        ∃ k, k ≤ example_dist p.loc 3 ∧ (forward_n example_rt k p).loc = 3 := by
        have wf := example_build_rt_wf
        obtain ⟨k, hk, H⟩ := bounded_hops_to_none_or_dst wf p
        let q := forward_n example_rt k p
        cases H with
        | inl hnone =>
          -- rt q.loc q.dst = none; show q.loc = 3 by analyzing example_neighbors
          have : rt q.loc q.dst = none := hnone
          dsimp [example_rt, build_rt, example_neighbors] at this
          -- case-split on q.loc; only q.loc = 3 yields none
          cases q.loc <;> simp [example_neighbors] at this
          any_goals simp at this
          -- after simplification, conclude q.loc = 3
          use k
          constructor
          · exact hk
          · simp [q]
        | inr heq =>
          use k
          constructor
          · exact hk
          · simp [q, heq, h]

/- Stronger loop-freedom: if `rt` is well-formed with respect to `dist`, then
   starting from any packet `p` there exists a bounded number of hops (≤ dist)
   after which routing yields no next-hop or reaches the destination. This is
   essentially a top-level wrapper around `bounded_hops_to_none_or_dst`. -/
theorem routing_loop_free_stronger {rt : RoutingTable} {dist : Distance}
    (wf : routing_wf rt dist) (p : Packet) :
  ∃ k, k ≤ dist p.loc p.dst ∧ (let q := forward_n rt k p; rt q.loc q.dst = none ∨ q.loc = q.dst) := by
  obtain ⟨k, hk, H⟩ := bounded_hops_to_none_or_dst wf p
  use k
  constructor
  · exact hk
  · exact H

/- If we have a numeric upper bound `m` such that `dist p.loc p.dst < m`, then
   the bounded-hops result yields an explicit `k < m` with the same guarantee. -/
theorem bounded_hops_bound_by_m {rt : RoutingTable} {dist : Distance}
    (wf : routing_wf rt dist) (p : Packet) (m : Nat) (h : dist p.loc p.dst < m) :
  ∃ k, k < m ∧ (let q := forward_n rt k p; rt q.loc q.dst = none ∨ q.loc = q.dst) := by
  obtain ⟨k, hk, H⟩ := bounded_hops_to_none_or_dst wf p
  have klt : k < m := lt_of_le_of_lt hk h
  use k
  constructor
  · exact klt
  · exact H

/- If `dist` is globally bounded by `n` for the destination `d` and `rt` is
   total toward `d`, then any packet destined for `d` reaches `d` within
   fewer than `n` hops. -/
theorem reach_dst_finite_nodes {rt : RoutingTable} {dist : Distance}
    (wf : routing_wf rt dist) {d : Node} (n : Nat)
    (dist_bound : ∀ x, dist x d < n)
    (total : ∀ x, x ≠ d → ∃ next, rt x d = some next)
    (p : Packet) (hp : p.dst = d) :
  ∃ k, k < n ∧ (forward_n rt k p).loc = d := by
  -- apply bounded_hops_bound_by_m with m = n
  have h0 : dist p.loc p.dst < n := by
    simp [hp]
    apply dist_bound
  obtain ⟨k, hk, H⟩ := bounded_hops_bound_by_m wf p n h0
  cases H with
  | inr heq =>
    use k; constructor; exact hk; simp [heq]
  | inl hnone =>
    let q := forward_n rt k p
    have hn : rt q.loc q.dst = none := hnone
    by_cases hq : q.loc = d
    · use k; constructor; exact hk; simp [hq]
    · -- q.loc ≠ d, but `total` gives a next-hop, contradicting `none`
      have ex := total q.loc (by intro Contra; apply hq; exact Contra.symm)
      obtain ⟨next, hnex⟩ := ex
      simp [hnex] at hn
      contradiction

end Smip

/- No cycles under `routing_wf`: following next-hops strictly decreases `dist`,
   so you cannot return to the same node along a sequence of next-hops. -/
theorem no_cycles_under_wf {rt : RoutingTable} {dist : Distance} (wf : routing_wf rt dist) :
  ∀ (p : Packet) (i k : Nat), k > 0 →
    (∀ m, i ≤ m → m < i + k → ∃ next, rt (forward_n rt m p).loc p.dst = some next) →
    (forward_n rt (i + k) p).loc ≠ (forward_n rt i p).loc := by
  intro p i k hk has_next
  -- show that for any n ≥ 1, dist at (i + n) is strictly smaller than dist at i
  have dist_decrease : ∀ n, 1 ≤ n → dist (forward_n rt (i + n) p).loc p.dst < dist (forward_n rt i p).loc p.dst := by
    intro n hn
    induction n with
    | zero => cases hn
    | succ n ih =>
      -- if n = 0 then succ n = 1, base case
      by_cases h0 : n = 0
      · -- base n = 0 (so original n = 1)
        have ex := has_next i (by simp) (by simp [Nat.lt_succ_self])
        obtain ⟨next, hnext⟩ := ex
        let p0 := forward_n rt i p
        have dlt := forward_step_decreases_dist wf p0
        exact dlt
      · -- n ≥ 1, so use IH for n and then one more step
        have npos : 1 ≤ n := by
          cases n
          · contradiction
          · simp
        have ih_res := ih npos
        -- apply forward_step_decreases_dist at position i + n
        have ex := has_next (i + n) (by apply Nat.le_add_right) (by
          calc
            i + n < i + (n + 1) := by apply Nat.add_lt_add_left (Nat.lt_succ_self _) i)
        obtain ⟨next, hnext⟩ := ex
        let p_n := forward_n rt (i + n) p
        have dlt := forward_step_decreases_dist wf p_n
        exact lt_trans dlt ih_res
  -- now pick n = k to derive a strict decrease from i to i+k
  have dec := dist_decrease k (by have : 1 ≤ k := by simp [hk]; exact this)
  intro contra
  -- if locations equal then distances equal, contradicting strict decrease
  have eq_dist : dist (forward_n rt (i + k) p).loc p.dst = dist (forward_n rt i p).loc p.dst := by
    -- when locations equal their distances to same dst are equal
    have eq_loc : (forward_n rt (i + k) p).loc = (forward_n rt i p).loc := contra
    simp [eq_loc]
  have impossible := lt_irrefl (dist (forward_n rt i p).loc p.dst) (lt_trans dec (Eq.symm eq_dist) )
  contradiction

/- If the routing table is well-formed and for a fixed destination `d` every
   non-destination node has a next-hop (i.e., `rt` is total toward `d`),
   then starting from any packet destined for `d` you will reach `d` within
   at most `dist p.loc d` hops. -/
theorem routing_reaches_dst_if_total {rt : RoutingTable} {dist : Distance}
    (wf : routing_wf rt dist) {d : Node}
    (total : ∀ n, n ≠ d → ∃ next, rt n d = some next)
    (p : Packet) (h : p.dst = d) :
  ∃ k, k ≤ dist p.loc d ∧ (forward_n rt k p).loc = d := by
  -- Use the bounded-hops theorem which gives either `rt q.loc q.dst = none` or `q.loc = q.dst`.
  obtain ⟨k, hk, H⟩ := bounded_hops_to_none_or_dst wf p
  cases H with
  | inl hnone =>
    let q := forward_n rt k p
    have hn : rt q.loc q.dst = none := hnone
    by_cases hq : q.loc = d
    · use k; constructor; exact hk; simp [hq]
    · -- q.loc ≠ d, but `total` yields some next-hop, contradicting `none`.
      have ex := total q.loc (by intro Contra; apply hq; exact Contra.symm)
      obtain ⟨next, hnex⟩ := ex
      simp [hnex] at hn
      contradiction
  | inr heq =>
    use k; constructor; exact hk; simp [heq, h]
