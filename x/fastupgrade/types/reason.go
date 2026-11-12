package types

// Reason is why the fast path did not start or fire. Its value is the
// `reason` event attribute. A check that returns an error returns the empty
// Reason, which is not ReasonOK.
type Reason string

const (
	// ReasonOK means that every check passed. Only the code uses it: no event
	// carries it.
	ReasonOK Reason = "ok"

	// Results of the support check: the validator gate, the gov
	// compatibility check, and gov's tally.
	ReasonIncompatibleWithGov Reason = "incompatible_with_gov"
	ReasonGateBelowThreshold  Reason = "gate_below_threshold"
	ReasonGovTallyRejected    Reason = "gov_tally_rejected"
	ReasonTallyBelowThreshold Reason = "tally_below_threshold"

	// Results of the checks that run before the support check.
	ReasonNaturalDeadline Reason = "natural_deadline"
	ReasonNotEligible     Reason = "not_eligible"
	ReasonPlanPending     Reason = "plan_pending"

	// Reasons that only a running timer can have.
	ReasonDisabled     Reason = "disabled"
	ReasonNotVoting    Reason = "not_voting"
	ReasonMissedWindow Reason = "missed_window"
	// ReasonInternalError is fixed text: the error itself is only logged,
	// because its text can differ between nodes.
	ReasonInternalError Reason = "internal_error"
)
