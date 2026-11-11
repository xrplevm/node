package types

import (
	"cosmossdk.io/math"
)

const (
	DefaultTimelockBlocks uint64 = 150
	MinTimelockBlocks     uint64 = 2
	MaxTimelockBlocks     uint64 = 100000

	DefaultPostponeBlocks uint64 = 10
	MinPostponeBlocks     uint64 = 1
	MaxPostponeBlocks     uint64 = 100000
)

// DefaultThreshold is 0.85 in gov's 18-decimal format.
var DefaultThreshold = math.LegacyNewDecWithPrec(85, 2)

// DefaultParams returns the disabled default parameters.
func DefaultParams() Params {
	return Params{
		Enabled:        false,
		TimelockBlocks: DefaultTimelockBlocks,
		Threshold:      DefaultThreshold.String(),
		PostponeBlocks: DefaultPostponeBlocks,
	}
}

// Validate checks the parameters without reading gov state.
func (p Params) Validate() error {
	if p.TimelockBlocks < MinTimelockBlocks || p.TimelockBlocks > MaxTimelockBlocks {
		return ErrInvalidParams.Wrapf("timelock_blocks must be in [%d, %d], got %d", MinTimelockBlocks, MaxTimelockBlocks, p.TimelockBlocks)
	}
	if p.PostponeBlocks < MinPostponeBlocks || p.PostponeBlocks > MaxPostponeBlocks {
		return ErrInvalidParams.Wrapf("postpone_blocks must be in [%d, %d], got %d", MinPostponeBlocks, MaxPostponeBlocks, p.PostponeBlocks)
	}
	_, err := p.ThresholdDec()
	return err
}

// ThresholdDec parses the threshold and checks it is in (0, 1].
func (p Params) ThresholdDec() (math.LegacyDec, error) {
	threshold, err := math.LegacyNewDecFromStr(p.Threshold)
	if err != nil {
		return math.LegacyDec{}, ErrInvalidParams.Wrapf("threshold %q: %s", p.Threshold, err)
	}
	if !threshold.IsPositive() || threshold.GT(math.LegacyOneDec()) {
		return math.LegacyDec{}, ErrInvalidParams.Wrapf("threshold must be in (0, 1], got %s", threshold)
	}
	return threshold, nil
}
