namespace Smip

/--
Lean model for the current Go routing policy surface.

This file captures the default policy seed and the exact-match update/lookup
behavior used by `internal/routing/router.go`.
-/

/-- Policy keys are modeled abstractly as natural numbers. -/
abbrev PolicyKey := Nat

/-- Lean counterpart of `RoutePolicy`. -/
structure RoutePolicy where
  NextHopID : Nat
  QueueID : Nat
  Priority : Nat
  deriving Repr, DecidableEq

/-- A simple policy table model. -/
abbrev PolicyTable := List (PolicyKey × RoutePolicy)

/-- The default policy inserted by `SeedDefaultPolicies`. -/
def defaultPolicy : RoutePolicy :=
  { NextHopID := 0, QueueID := 0, Priority := 10 }

/-- The seeded router contains the default policy at key `0`. -/
def seedDefaultPolicies : PolicyTable :=
  [(0, defaultPolicy)]

/-- Exact-match policy lookup with first-match semantics. -/
def lookupPolicy : PolicyTable → PolicyKey → Option RoutePolicy
  | [], _ => none
  | (k, p) :: rest, key => if k = key then some p else lookupPolicy rest key

/-- Updating a policy is modeled as inserting the new key at the front. -/
def updatePolicy (policies : PolicyTable) (key : PolicyKey) (policy : RoutePolicy) : PolicyTable :=
  (key, policy) :: policies

/-- Minimal additive routing wrapper model for enhanced routing behavior. -/
structure LPMEntry where
  Prefix : ByteSeq
  PrefixLen : Nat
  NextHopID : ByteSeq
  Metric : Nat
  deriving Repr, DecidableEq

/-- Enhanced routing keeps the base table and adds LPM entries additively. -/
structure EnhancedRouteTable where
  base : RouteTable
  lpmEntries : List LPMEntry
  deriving Repr, DecidableEq

/-- Add a longest-prefix-match route without mutating the base exact-match table. -/
def addLPMRoute (table : EnhancedRouteTable) (entry : LPMEntry) : EnhancedRouteTable :=
  { table with lpmEntries := table.lpmEntries ++ [entry] }

/-- Enhanced routing is additive: the base exact-match table is unchanged. -/
theorem addLPMRoute_preserves_base (table : EnhancedRouteTable) (entry : LPMEntry) :
    (addLPMRoute table entry).base = table.base := by
  rfl

/-- Lookup with the seeded default policy as a fallback. -/
def lookupPolicyOrDefault (policies : PolicyTable) (key : PolicyKey) : RoutePolicy :=
  match lookupPolicy policies key with
  | some policy => policy
  | none => defaultPolicy

/-- The seeded router always resolves key `0` to the default policy. -/
theorem lookup_seed_default : lookupPolicy seedDefaultPolicies 0 = some defaultPolicy := by
  simp [seedDefaultPolicies, lookupPolicy, defaultPolicy]

/-- The seeded router has no exact policy for nonzero keys. -/
theorem lookup_seed_default_other (k : PolicyKey) (hk : k ≠ 0) :
    lookupPolicy seedDefaultPolicies k = none := by
  simp [seedDefaultPolicies, lookupPolicy, defaultPolicy, hk]

/-- Updating a policy makes the new policy visible to exact lookup at the same key. -/
theorem lookup_updatePolicy_hit (policies : PolicyTable) (key : PolicyKey) (policy : RoutePolicy) :
    lookupPolicy (updatePolicy policies key policy) key = some policy := by
  simp [updatePolicy, lookupPolicy]

/-- Updating one key does not affect exact lookup for a different key. -/
theorem lookup_updatePolicy_other (policies : PolicyTable) (key otherKey : PolicyKey)
    (policy : RoutePolicy) (hne : otherKey ≠ key) :
    lookupPolicy (updatePolicy policies key policy) otherKey = lookupPolicy policies otherKey := by
  simp [updatePolicy, lookupPolicy, hne]

/-- The default fallback is used when no exact policy exists. -/
theorem lookupPolicyOrDefault_none (policies : PolicyTable) (key : PolicyKey)
    (h : lookupPolicy policies key = none) :
    lookupPolicyOrDefault policies key = defaultPolicy := by
  simp [lookupPolicyOrDefault, h, defaultPolicy]

/-- When exact lookup succeeds, the default wrapper returns that policy unchanged. -/
theorem lookupPolicyOrDefault_hit (policies : PolicyTable) (key : PolicyKey)
    (policy : RoutePolicy) (h : lookupPolicy policies key = some policy) :
    lookupPolicyOrDefault policies key = policy := by
  simp [lookupPolicyOrDefault, h]

/-- A compact correctness statement for default-or-exact policy lookup. -/
theorem lookupOrPredict_policy_correct (policies : PolicyTable) (key : PolicyKey) :
    match lookupPolicy policies key with
    | some policy => lookupPolicyOrDefault policies key = policy
    | none => lookupPolicyOrDefault policies key = defaultPolicy := by
  cases h : lookupPolicy policies key <;> simp [lookupPolicyOrDefault, h, defaultPolicy]

/-- The enhanced routing wrapper leaves the base exact lookup behavior unchanged. -/
theorem enhanced_additive_wrapper_preserves_lookup (table : EnhancedRouteTable) (entry : LPMEntry)
    (dst : ByteSeq) :
    lookupNextHop (addLPMRoute table entry).base dst = lookupNextHop table.base dst := by
  simp [addLPMRoute]

/-- Updating a policy keeps the new priority visible to lookup. -/
theorem updatePolicy_priority_one (policies : PolicyTable) (key : PolicyKey)
    (policy : RoutePolicy) :
    (lookupPolicyOrDefault (updatePolicy policies key { policy with Priority := 1 }) key).Priority = 1 := by
  simp [lookupPolicyOrDefault, lookup_updatePolicy_hit, updatePolicy]

/-- The seeded default policy carries the expected lower-priority value. -/
theorem defaultPolicy_priority_ten : defaultPolicy.Priority = 10 := by
  rfl

end Smip