package types

import (
	"slices"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultTimelockBlocks uint64 = 150
	MinTimelockBlocks     uint64 = 2
	MaxTimelockBlocks     uint64 = 100000

	DefaultPostponeBlocks uint64 = 10
	MinPostponeBlocks     uint64 = 1
	MaxPostponeBlocks     uint64 = 100000

	// MaxProposers limits the proposers list. Validate runs on every
	// MsgUpdateParams that a node receives, before the signature and the fee
	// are checked, and it compares each proposer with all earlier ones. With
	// no limit, anyone could send long lists at no cost and make nodes spend
	// CPU on them. The vote hook also reads the params on every vote.
	//
	// 16 is not a tuned value: any small number keeps the check cheap. The
	// v12 upgrade stores one proposer, so 16 leaves room for more.
	MaxProposers = 16
)

// DefaultThreshold is 0.85 in gov's 18-decimal format.
var DefaultThreshold = math.LegacyNewDecWithPrec(85, 2)

// DefaultParams returns the disabled default parameters. The proposers list
// is empty, so no proposal can use the fast path until gov adds a proposer.
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
	if len(p.Proposers) > MaxProposers {
		return ErrInvalidParams.Wrapf("proposers must have at most %d entries, got %d", MaxProposers, len(p.Proposers))
	}
	for i, proposer := range p.Proposers {
		// The fast path compares the stored text with gov's proposer, so the
		// address must be in the form that gov stores.
		addr, err := sdk.AccAddressFromBech32(proposer)
		if err != nil || addr.String() != proposer {
			return ErrInvalidParams.Wrapf("proposers: %q is not a canonical account address", proposer)
		}
		if slices.Contains(p.Proposers[:i], proposer) {
			return ErrInvalidParams.Wrapf("proposers: duplicate %s", proposer)
		}
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
