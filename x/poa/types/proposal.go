package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeAddValidator    = "PoaAddValidator"
	ProposalTypeRemoveValidator = "PoaRemoveValidator"
)

// Assert MsgAddValidator implements govtypes.Content at compile-time
var _ govtypes.Content = &AddValidatorProposal{}
var _ govtypes.Content = &RemoveValidatorProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddValidator)
	govtypes.RegisterProposalType(ProposalTypeRemoveValidator)
}

func (msg *AddValidatorProposal) ProposalRoute() string { return RouterKey }
func (msg *AddValidatorProposal) ProposalType() string  { return ProposalTypeAddValidator }
func (msg *AddValidatorProposal) ValidateBasic() error {
	// TODO: Add basic address validation
	return govtypes.ValidateAbstract(msg)
}
func NewAddValidatorProposal(title string, description string, validatorAddress string) govtypes.Content {
	return &AddValidatorProposal{Title: title, Description: description, ValidatorAddress: validatorAddress}
}

func (msg *RemoveValidatorProposal) ProposalRoute() string { return RouterKey }
func (msg *RemoveValidatorProposal) ProposalType() string  { return ProposalTypeAddValidator }
func (msg *RemoveValidatorProposal) ValidateBasic() error {
	// TODO: Add basic address validation
	return govtypes.ValidateAbstract(msg)
}
func NewRemoveValidatorProposal(title string, description string, validatorAddress string) govtypes.Content {
	return &RemoveValidatorProposal{Title: title, Description: description, ValidatorAddress: validatorAddress}
}
