package utils

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

type ValidatorFlagOverride struct {
	Index int
	Flag  cmtproto.BlockIDFlag
}

func NewValidatorFlagOverride(index int, flag cmtproto.BlockIDFlag) ValidatorFlagOverride {
	return ValidatorFlagOverride{
		Index: index,
		Flag:  flag,
	}
}

func NewValidatorFlags(n int, overrides ...ValidatorFlagOverride) []cmtproto.BlockIDFlag {
	flags := make([]cmtproto.BlockIDFlag, n)
	for i := range flags {
		flags[i] = cmtproto.BlockIDFlagCommit
	}

	for _, override := range overrides {
		flags[override.Index] = override.Flag
	}

	return flags
}
