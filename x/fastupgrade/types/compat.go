package types

import (
	"cosmossdk.io/math"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// CheckGovCompatibility reports whether threshold can never accelerate a
// proposal that gov's own tally would reject. LegacyDec values are integers
// scaled by 10^18, so every comparison is exact.
func CheckGovCompatibility(threshold math.LegacyDec, gov govv1.Params) error {
	govThreshold, err := math.LegacyNewDecFromStr(gov.Threshold)
	if err != nil {
		return ErrIncompatibleWithGov.Wrapf("gov threshold: %s", err)
	}
	expedited, err := math.LegacyNewDecFromStr(gov.ExpeditedThreshold)
	if err != nil {
		return ErrIncompatibleWithGov.Wrapf("gov expedited_threshold: %s", err)
	}
	quorum, err := math.LegacyNewDecFromStr(gov.Quorum)
	if err != nil {
		return ErrIncompatibleWithGov.Wrapf("gov quorum: %s", err)
	}
	veto, err := math.LegacyNewDecFromStr(gov.VetoThreshold)
	if err != nil {
		return ErrIncompatibleWithGov.Wrapf("gov veto_threshold: %s", err)
	}

	switch {
	case !threshold.GT(govThreshold):
		return ErrIncompatibleWithGov.Wrapf("threshold %s must be greater than gov threshold %s", threshold, govThreshold)
	case !threshold.GT(expedited):
		return ErrIncompatibleWithGov.Wrapf("threshold %s must be greater than gov expedited_threshold %s", threshold, expedited)
	case threshold.LT(quorum):
		return ErrIncompatibleWithGov.Wrapf("threshold %s must be at least gov quorum %s", threshold, quorum)
	case threshold.LT(math.LegacyOneDec().Sub(veto)):
		return ErrIncompatibleWithGov.Wrapf("threshold %s must be at least 1 - gov veto_threshold %s", threshold, veto)
	}
	return nil
}
