package types

// Reason is why the fast path did not start or fire. Its value is the
// `reason` event attribute. A check that returns an error returns the empty
// Reason, which is not ReasonOK.
type Reason string

const (
	ReasonOK Reason = "ok"

	ReasonIncompatibleWithGov Reason = "incompatible-with-gov"
	ReasonGateBelowThreshold  Reason = "gate-below-threshold"
	ReasonGovTallyRejected    Reason = "gov-tally-rejected"
	ReasonTallyBelowThreshold Reason = "tally-below-threshold"

	ReasonNaturalDeadline Reason = "natural-deadline"
	ReasonNotMinor        Reason = "not-minor"
	ReasonPlanPending     Reason = "plan-pending"

	ReasonDisabled     Reason = "disabled"
	ReasonNotVoting    Reason = "not-voting"
	ReasonMissedWindow Reason = "missed-window"
	// ReasonInternalError is fixed text: the error itself is only logged,
	// because its text can differ between nodes.
	ReasonInternalError Reason = "internal-error"
)
