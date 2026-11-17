package types

const (
	EventTypeStarted    = "fastupgrade_started"
	EventTypeNotStarted = "fastupgrade_not_started"
	EventTypeDropped    = "fastupgrade_dropped"
	EventTypePostponed  = "fastupgrade_postponed"
	EventTypeEarlyClose = "fastupgrade_early_close"

	AttributeKeyProposalID    = "proposal_id"
	AttributeKeyPlanName      = "plan_name"
	AttributeKeyPlanInfo      = "plan_info"
	AttributeKeyUpgradeHeight = "upgrade_height"
	AttributeKeyReason        = "reason"
)
