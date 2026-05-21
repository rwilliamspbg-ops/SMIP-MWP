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

/-- Updating a policy keeps the new priority visible to lookup. -/
theorem updatePolicy_priority_one (policies : PolicyTable) (key : PolicyKey)
    (policy : RoutePolicy) :
    (lookupPolicyOrDefault (updatePolicy policies key { policy with Priority := 1 }) key).Priority = 1 := by
  simp [lookupPolicyOrDefault, lookup_updatePolicy_hit, updatePolicy]

/-- The seeded default policy carries the expected lower-priority value. -/
theorem defaultPolicy_priority_ten : defaultPolicy.Priority = 10 := by
  rfl

end Smip